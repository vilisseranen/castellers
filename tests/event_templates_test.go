package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/vilisseranen/castellers/model"
)

const weeklyPractice = `{
	"title": " Entraînement hebdomadaire ",
	"name": "Entraînement",
	"type": "practice",
	"description": "Apporte ta faixa et de l'eau.",
	"locationName": "Collège Brébeuf",
	"location": {"lat": 45.5007, "lng": -73.6241},
	"startDate": 1794523500,
	"endDate": 1794533400,
	"uniformRequired": 0}`

func (test *TestHelper) eventTemplates(t *testing.T, adminToken string) []model.EventTemplate {
	t.Helper()
	code, body := test.duesRequest("GET", "/event-templates", adminToken, nil)
	if err := test.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	var templates []model.EventTemplate
	json.Unmarshal(body, &templates)
	return templates
}

func TestEventTemplates(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()

	if templates := h.eventTemplates(t, adminToken); len(templates) != 0 {
		t.Fatalf("Expected no template. Got %d", len(templates))
	}

	// Create.
	code, body := h.duesRequest("POST", "/event-templates", adminToken, []byte(weeklyPractice))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	var created model.EventTemplate
	json.Unmarshal(body, &created)
	if created.UUID == "" || created.Title != "Entraînement hebdomadaire" {
		t.Errorf("Unexpected created template: %+v", created)
	}

	templates := h.eventTemplates(t, adminToken)
	if len(templates) != 1 {
		t.Fatalf("Expected 1 template. Got %d", len(templates))
	}
	got := templates[0]
	if got.UUID != created.UUID || got.Name != "Entraînement" || got.Type != "practice" ||
		got.Description != "Apporte ta faixa et de l'eau." || got.LocationName != "Collège Brébeuf" ||
		got.Location.Lat != 45.5007 || got.Location.Lng != -73.6241 ||
		got.StartDate != 1794523500 || got.EndDate != 1794533400 || got.UniformRequired != 0 {
		t.Errorf("Unexpected template: %+v", got)
	}

	// A second one; the list is sorted by title, whatever the case.
	code, _ = h.duesRequest("POST", "/event-templates", adminToken,
		[]byte(`{"title":"diada","name":"Diada","type":"presentation","uniformRequired":1}`))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	templates = h.eventTemplates(t, adminToken)
	if len(templates) != 2 || templates[0].Title != "diada" || templates[1].Title != "Entraînement hebdomadaire" {
		t.Fatalf("Expected 'diada' then 'Entraînement hebdomadaire'. Got %+v", templates)
	}
	if templates[0].UniformRequired != 1 || templates[0].Description != "" {
		t.Errorf("Unexpected second template: %+v", templates[0])
	}

	// Update: the content changes, the UUID stays.
	code, _ = h.duesRequest("PUT", "/event-templates/"+created.UUID, adminToken,
		[]byte(`{"title":"Entraînement hebdomadaire","name":"Entraînement du jeudi","type":"practice","description":"Nouvelle description","startDate":1794523500,"endDate":1794533400}`))
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	templates = h.eventTemplates(t, adminToken)
	if templates[1].UUID != created.UUID || templates[1].Name != "Entraînement du jeudi" || templates[1].Description != "Nouvelle description" || templates[1].LocationName != "" {
		t.Errorf("Unexpected updated template: %+v", templates[1])
	}

	// Delete.
	code, _ = h.duesRequest("DELETE", "/event-templates/"+created.UUID, adminToken, nil)
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Fatal(err)
	}
	if templates := h.eventTemplates(t, adminToken); len(templates) != 1 || templates[0].Title != "diada" {
		t.Errorf("Expected only 'diada' to be left. Got %+v", templates)
	}
	code, _ = h.duesRequest("DELETE", "/event-templates/"+created.UUID, adminToken, nil)
	if err := h.checkResponseCode(http.StatusNotFound, code); err != nil {
		t.Error(err)
	}
	code, _ = h.duesRequest("PUT", "/event-templates/"+created.UUID, adminToken, []byte(weeklyPractice))
	if err := h.checkResponseCode(http.StatusNotFound, code); err != nil {
		t.Error(err)
	}
}

func TestEventTemplatesValidation(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()

	// Templates are for admins only.
	for _, request := range [][2]string{{"GET", "/event-templates"}, {"POST", "/event-templates"}, {"PUT", "/event-templates/0123"}, {"DELETE", "/event-templates/0123"}} {
		code, _ := h.duesRequest(request[0], request[1], memberToken, []byte(weeklyPractice))
		if err := h.checkResponseCode(http.StatusUnauthorized, code); err != nil {
			t.Errorf("%s %s: %v", request[0], request[1], err)
		}
	}

	invalid := map[string]string{
		"no title":       `{"title":"  ","type":"practice"}`,
		"unknown type":   `{"title":"Sortie","type":"party"}`,
		"no type":        `{"title":"Sortie"}`,
		"end before":     `{"title":"Sortie","type":"social","startDate":200,"endDate":100}`,
		"uniform":        `{"title":"Sortie","type":"social","uniformRequired":2}`,
		"not json":       `nope`,
		"title too long": `{"title":"` + longTitle() + `","type":"social"}`,
	}
	for name, payload := range invalid {
		code, _ := h.duesRequest("POST", "/event-templates", adminToken, []byte(payload))
		if err := h.checkResponseCode(http.StatusBadRequest, code); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if templates := h.eventTemplates(t, adminToken); len(templates) != 0 {
		t.Fatalf("Expected no template. Got %d", len(templates))
	}

	// Two templates cannot have the same title, whatever the case.
	code, body := h.duesRequest("POST", "/event-templates", adminToken, []byte(weeklyPractice))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	var first model.EventTemplate
	json.Unmarshal(body, &first)
	code, _ = h.duesRequest("POST", "/event-templates", adminToken, []byte(`{"title":"entraînement hebdomadaire","type":"practice"}`))
	if err := h.checkResponseCode(http.StatusConflict, code); err != nil {
		t.Error(err)
	}
	// Saving a template again under its own title is fine.
	code, _ = h.duesRequest("PUT", "/event-templates/"+first.UUID, adminToken, []byte(weeklyPractice))
	if err := h.checkResponseCode(http.StatusOK, code); err != nil {
		t.Error(err)
	}
	// Renaming one to the title of another is not.
	code, body = h.duesRequest("POST", "/event-templates", adminToken, []byte(`{"title":"Diada","type":"presentation"}`))
	if err := h.checkResponseCode(http.StatusCreated, code); err != nil {
		t.Fatal(err)
	}
	var second model.EventTemplate
	json.Unmarshal(body, &second)
	code, _ = h.duesRequest("PUT", "/event-templates/"+second.UUID, adminToken, []byte(`{"title":"Entraînement hebdomadaire","type":"presentation"}`))
	if err := h.checkResponseCode(http.StatusConflict, code); err != nil {
		t.Error(err)
	}
}

func longTitle() string {
	title := make([]byte, 101)
	for i := range title {
		title[i] = 'a'
	}
	return string(title)
}
