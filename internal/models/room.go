package models

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/hngprojects/hng_boilerplate_golang_web/pkg/repository/storage/postgresql"
)

type Room struct {
	ID          string    `gorm:"type:uuid;primary_key" json:"room_id"`
	Name        string    `gorm:"column:name;unique type:text; not null" json:"name"`
	Description string    `gorm:"column:description; type:text; not null" json:"description"`
	Users       []User    `gorm:"many2many:user_rooms;" json:"users"`
	CreatedAt   time.Time `gorm:"column:created_at; not null; autoCreateTime" json:"created_at"`
	DeletedAt   time.Time `gorm:"column: deleted_at; not null; autoDeleteTime" json:"deleted_at"`
}

type UserRoom struct {
	RoomID    string    `gorm:"type:uuid;primaryKey;not null" json:"room_id"`
	UserID    string    `gorm:"type:uuid;primaryKey;not null" json:"user_id"`
	Username  string    `gorm:"column:username; type:varchar(255)" json:"username"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	DeletedAt time.Time `gorm:"index" json:"deleted_at"`
}

type CreateRoomRequest struct {
	Username    string `json:"username" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type GetRoomRequest struct {
	Name string `json:"name" validate:"required"`
}

type JoinRoomRequest struct {
	Username string `json:"username" validate:"required"`
	RoomID   string `json:"room_id" `
	UserID   string `json:"user_id" `
}

type UpdateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *Room) CreateRoom(db *gorm.DB) error {
	err := postgresql.CreateOneRecord(db, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *Room) GetUsersInRoomByID(db *gorm.DB, roomID string) ([]UserRoom, error) {
	var users []UserRoom

	err := postgresql.SelectUsersFromDb(
		db.Where("room_id = ?", roomID),
		"",
		&users,
		"room_id = ?",
		roomID,
	)

	if err != nil {
		return users, err
	}

	return users, nil
}

func (r *Room) GetRoomByID(db *gorm.DB, roomID string) (Room, error) {
	var room Room

	err, _ := postgresql.SelectOneFromDb(db, &room, "room_id= ?", roomID)
	if err != nil {
		return room, errors.New("room not found")
	}
	return room, nil
}

func (r *Room) GetRoomByName(db *gorm.DB, name string) (Room, error) {
	var room Room

	exists := postgresql.CheckExists(db, &room, "name= ?", name)
	if !exists {
		return room, errors.New("room not found")
	}

	err, _ := postgresql.SelectOneFromDb(db, &room, "name= ?", name)
	if err != nil {
		return room, err
	}
	return room, nil
}

func (r *Room) GetRooms(db *gorm.DB) ([]Room, error) {
	var rooms []Room

	err := postgresql.SelectAllFromDb(db, "", &rooms, "")
	if err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (r *Room) GetRoomMessages(db *gorm.DB, userID, roomID string) ([]Message, error) {

	var (
		messages []Message
		userRoom UserRoom
	)

	exist := postgresql.CheckExists(db, &userRoom, "room_id = ? AND user_id = ?", roomID, userID)
	if !exist {
		return messages, errors.New("user not in room")
	}

	err := postgresql.SelectAllFromDb(
		db.Where("room_id = ?", roomID),
		"",
		&messages,
		"room_id = ?",
		roomID,
	)
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func (r *Room) AddUserToRoom(db *gorm.DB, req JoinRoomRequest) error {

	var (
		user   User
		room   Room
		userID = req.UserID
		roomID = req.RoomID
	)

	exists := postgresql.CheckExists(db, &user, "user_id = ?", userID)
	if !exists {
		return errors.New("user does not exist")
	}

	exists = postgresql.CheckExists(db, &room, "room_id = ?", roomID)
	if !exists {
		return errors.New("room does not exist")
	}

	var userRoom UserRoom
	exist := postgresql.CheckExists(db, &userRoom, "room_id = ? AND user_id = ?", roomID, userID)
	if exist {
		return errors.New("user already in room")
	}

	userRoom = UserRoom{
		RoomID:   roomID,
		UserID:   userID,
		Username: req.Username,
	}

	err := postgresql.CreateOneRecord(db, &userRoom)
	if err != nil {
		return errors.New("could not add user to room")
	}
	return nil
}

func (r *Room) RemoveUserFromRoom(db *gorm.DB, roomID, userID string) error {
	var userRoom UserRoom

	exist := postgresql.CheckExists(db, &userRoom, "room_id = ? AND user_id = ?", roomID, userID)
	if !exist {
		return errors.New("user not in room")
	}

	err := postgresql.DeleteRecordFromDb(db, &userRoom)
	if err != nil {
		return errors.New("could not remove user from room")
	}
	return nil
}

func (u *UserRoom) CountRoomUsers(db *gorm.DB, roomID string) (int, error) {
	var users []UserRoom

	exists := postgresql.CheckExists(db, &u, "room_id = ?", roomID)
	if !exists {
		return 0, errors.New("room does not exist")
	}

	err := postgresql.SelectAllFromDb(db, "", &users, "room_id = ?", roomID)
	if err != nil {
		return 0 , err
	}

	return len(users), nil
}

func (r *Room) UpdateRoom(db *gorm.DB, req UpdateRoomRequest, roomID string) (Room, int, error) {
	var room Room
	room.ID = roomID

	exists := postgresql.CheckExists(db, &room, "room_id = ?", roomID)
	if !exists {
		return room, http.StatusNotFound, errors.New("room does not exist")
	}

	room.Name = req.Name
	room.Description = req.Description

	_, err := postgresql.SaveAllFields(db, room)
	if err != nil {
		return room, http.StatusInternalServerError, nil
	}

	updatedRoom := Room{}
	err = db.First(&updatedRoom, "room_id = ?", roomID).Error
	if err != nil {
		return room, http.StatusInternalServerError, err
	}
	return updatedRoom, http.StatusOK, nil
}
