package main

import "github.com/abankelsey/study_helper/internal/data"

type TemplateData struct {
	Title       string
	HeaderText  string
	FormErrors  map[string]string
	FormData    map[string]string
	GoalList    []*data.Goals    //stores the list of goal entries
	SessionList []*data.Sessions //stores the list of session entries

}

func NewTemplateData() *TemplateData {
	return &TemplateData{
		Title:       "Default Title",
		HeaderText:  "Default HeaderText",
		FormErrors:  map[string]string{},
		FormData:    map[string]string{},
		GoalList:    []*data.Goals{},    // Initialize the list as an empty slice
		SessionList: []*data.Sessions{}, // Initialize the list as an empty slice

	}
}
