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
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/bufferflies/pd-analyze/repository"
)

type ProjectServer struct {
	project repository.ProjectStorage
}

func NewProjectServer(storage repository.ProjectStorage) *ProjectServer {
	return &ProjectServer{
		project: storage,
	}
}

func (s *ProjectServer) NewProject(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	description := query.Get("description")
	err := s.project.Save(name, description)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *ProjectServer) GetProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.project.GetAll()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	body, err := json.Marshal(projects)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(body))
}

func (s *ProjectServer) NewSession(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pid, err := strconv.ParseUint(query.Get("project_id"), 10, 10)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	name := query.Get("name")
	targetObject := query.Get("target_object")
	objects := query["objects"]
	err = s.project.SaveSession(uint(pid), name, targetObject, objects)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *ProjectServer) GetSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["project_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	sessions, err := s.project.GetSessions(uint(pid))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	body, err := json.Marshal(sessions)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(body))
}

func (s *ProjectServer) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["session_id"], 10, 32)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	session, err := s.project.GetSession(uint(sid))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	body, err := json.Marshal(session)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(body))
}

func (s *ProjectServer) DeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["session_id"], 10, 32)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	if err = s.project.DeleteSession(uint(sid)); err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "ok")
}
