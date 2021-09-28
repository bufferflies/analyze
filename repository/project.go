// Copyright 2021 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package repository

import (
	"strings"

	"gorm.io/gorm"
)

type ProjectStorage interface {
	Save(name, description string) error
	SaveSession(session Session) error
	UpdateSession(sid uint, name, targetObject string, object []string) error

	GetAll() ([]Project, error)
	GetSessions(projectID uint) ([]Session, error)
	GetSession(sessionID uint) (Session, error)

	DeleteProject(projectID uint) error
	DeleteSession(sessionID uint) error
}

type ProjectDao struct {
	db *gorm.DB
}

func NewProjectDao(db *gorm.DB) ProjectStorage {
	db.AutoMigrate(&Project{}, &Session{})
	return &ProjectDao{db: db}
}

func (p ProjectDao) Save(name, description string) error {
	m := p.db.Create(&Project{Name: name, Description: description})
	return m.Error
}

func (p ProjectDao) SaveSession(session Session) error {
	if session.ID > 0 {
		if m := p.db.Model(&session).Updates(&session); m.Error != nil {
			return m.Error
		}
	} else {
		if m := p.db.Save(&session); m.Error != nil {
			return m.Error
		}
	}
	return nil
}

func (p ProjectDao) UpdateSession(sid uint, name, targetObject string, object []string) error {
	objects := strings.Join(object, ",")
	m := p.db.Model(&Session{}).Updates(&Session{ID: sid, Name: name, Object: objects, TargetObject: targetObject})
	return m.Error
}

func (p ProjectDao) GetAll() ([]Project, error) {
	var projects []Project
	m := p.db.Find(&projects)
	return projects, m.Error
}

func (p ProjectDao) GetSession(sessionID uint) (Session, error) {
	var session Session
	m := p.db.Where(&Session{ID: sessionID}).Find(&session)
	return session, m.Error
}

func (p ProjectDao) GetSessions(projectID uint) ([]Session, error) {
	var sessions []Session
	m := p.db.Where(&Session{PID: projectID}).Find(&sessions)
	return sessions, m.Error
}

func (p ProjectDao) DeleteProject(projectID uint) error {
	m := p.db.Delete(&Project{ID: projectID})
	if m.Error != nil {
		return m.Error
	}
	m = p.db.Delete(&Session{PID: projectID})
	return m.Error
}

func (p ProjectDao) DeleteSession(sessionID uint) error {
	if m := p.db.Where(&Session{ID: sessionID}).Delete(&Session{}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Where(&Workload{SessionID: sessionID}).Delete(&Workload{}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Where(&Metrics{SessionID: sessionID}).Delete(&Metrics{}); m.Error != nil {
		return m.Error
	}
	return nil
}
