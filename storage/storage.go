package storage

import "errors"

type User struct {
	Name         string
	ClassName    string
	RefreshToken string
	AccessToken  string
}

type Storage interface {
	AddUser(user *User) error
	GetUserByName(name string) (*User, bool)
	GetUserByRefreshToken(refreshToken string) (*User, bool)
	GetUserByAccessToken(accessToken string) (*User, bool)
}

type memoryStorage struct {
	data map[string]*User
}

func NewMemoryStorage() Storage {
	return memoryStorage{
		data: make(map[string]*User),
	}
}

func (storage memoryStorage) AddUser(user *User) error {
	var _, exists = storage.data[user.Name]
	if exists {
		return errors.New("User already exists")
	}
	storage.data[user.Name] = user
	return nil
}

func (storage memoryStorage) GetUserByName(name string) (*User, bool) {
	var user, exists = storage.data[name]
	return user, exists
}

func (storage memoryStorage) GetUserByRefreshToken(refreshToken string) (*User, bool) {
	for _, user := range storage.data {
		if user.RefreshToken == refreshToken {
			return user, true
		}
	}
	return nil, false
}

func (storage memoryStorage) GetUserByAccessToken(accessToken string) (*User, bool) {
	for _, user := range storage.data {
		if user.AccessToken == accessToken {
			return user, true
		}
	}
	return nil, false
}
