package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pprisn/go-contacts/app"
	"github.com/pprisn/go-contacts/controllers"
)

func main() {

	//Определим объект маршрутов
	router := mux.NewRouter()
	//Определим обработчики маршрутов
	router.HandleFunc("/api/user/new", controllers.CreateAccount).Methods("POST")
	router.HandleFunc("/api/user/login", controllers.Authenticate).Methods("POST")
	router.HandleFunc("/api/contacts/new", controllers.CreateContact).Methods("POST")
	router.HandleFunc("/api/me/contacts", controllers.GetContactsFor).Methods("GET") //  user/2/contacts
	router.HandleFunc("/api/admin/users", controllers.GetUsers).Methods("GET")
	router.HandleFunc("/api/admin/update/users", controllers.UpdateUsers).Methods("POST") //служебная функция для административных изменений данных списка учетных записей

	//Добавим требование запуска проверки middleware для объектов обработки маршрутов !
	router.Use(app.JwtAuthentication) //attach JWT auth middleware

	//Заглушка для не существующего маршрута !
	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //localhost
	}

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}

}
