package service

import (
	"context"
	"log"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hako/branca"
	"golang.org/x/crypto/bcrypt"
)

func TestShouldReturnSomething(t *testing.T) {

	// Creates sqlmock database connection and a mock to manage expectations.
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// Closes the database and prevents new queries from starting.
	defer db.Close()

	hPassword, err := bcrypt.GenerateFromPassword([]byte("eiuwfbwieubfiuwbef"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("apple@gmail.com", "wdwdewffwef").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO credentials").WithArgs(0, hPassword).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.TODO()

	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	origin, err := url.Parse("http://localhost:4000")
	if err != nil || !origin.IsAbs() {
		log.Fatalf("invalid origin url: %v\n", err)
		return
	}

	s := New(db, codec, *origin)

	// Calls MenuByNameAndLanguage with mocked database connection in arguments list.
	err = s.CreateUser(ctx, "apple@gmail.com", "wdwdewffwef", "eiuwfbwieubfiuwbef")
	log.Println(err)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	//// Here we just construction our expecting result.
	//var menuLinks []models.MenuLink
	//menuLink1 := models.MenuLink{
	//	ID:       1,
	//	Title:    "enTitle",
	//	Langcode: "en",
	//	URL:      "/en-link",
	//	SideMenu: "0",
	//}
	//menuLinks = append(menuLinks, menuLink1)
	//
	//menuLink2 := models.MenuLink{
	//	ID:       2,
	//	Title:    "enTitle2",
	//	Langcode: "en",
	//	URL:      "/en-link2",
	//	SideMenu: "0",
	//}
	//
	//menuLinks = append(menuLinks, menuLink2)
	//
	//expectedMenu := &models.Menu{
	//	Name:  "main",
	//	Links: menuLinks,
	//}
	//
	//// And, finally, let's check if result from MenuByNameAndLanguage equal with expected result.// Here I used Testify library (https://github.com/stretchr/testify).
	//assert.Equal(t, expectedMenu, menu)
}

