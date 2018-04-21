package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

type Person struct {
	ID          string `form:"ID" storm:"id,increment" json:"ID"`
	User        string
	Name        string `form:"Name" storm:"index" json:"Name"`
	NameService string `form:"NameService" storm:"index" json:"NameService"`
	Date        string `form:"Date" storm:"index" json:"Date"`
	Number      string `form:"Number" storm:"index" json:"Number"`
}

//TODO
// type myForm struct {
// 	Admins
// 	Colors []string `form:"colors[]"`
// }

func SetupRouter() *gin.Engine {
	g := gin.Default()

	// v1 := router.Group("api/v1")
	// {
	// 	v1.GET("/instructions", GetInstructions)
	// }

	return g
}

func main() {

	//ADD EXAMPLE BOLTDB
	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	//router = gin.Default()

	db, err := storm.Open("db/data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// g := gin.New()
	g := SetupRouter()

	g.LoadHTMLGlob("templates/*.html")

	perm, err := permissionbolt.New()
	if err != nil {
		log.Fatalln(err)
	}

	// Blank slate, no default permissions
	//perm.Clear()

	// Set up a middleware handler for Gin, with a custom "permission denied" message.
	permissionHandler := func(c *gin.Context) {
		// Check if the user has the right admin/user rights
		if perm.Rejected(c.Writer, c.Request) {
			// Deny the request, don't call other middleware handlers
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
			return
		}
		// Call the next middleware handler
		c.Next()
	}

	// Logging middleware
	g.Use(gin.Logger())

	// Enable the permissionbolt middleware, must come before recovery
	g.Use(permissionHandler)

	// Recovery middleware
	g.Use(gin.Recovery())

	// Get the userstate, used in the handlers below
	userstate := perm.UserState()

	isloggedin := func(c *gin.Context) bool {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		return isloggedin
	}

	g.GET("/", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		if isloggedin {
			c.HTML(http.StatusOK, "index.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.Redirect(307, "/login")
		}
	})

	// Registaration Users GET
	g.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	// Registaration Users OPOST
	g.POST("/register", func(c *gin.Context) {

		username := c.PostForm("username")
		pass := c.PostForm("password")
		message := c.PostForm("email")

		userstate.AddUser(username, pass, message)
		userstate.Login(c.Writer, username)
		userstate.MarkConfirmed(username)

		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	// Loging Users GET
	g.GET("/login", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Page", "is_logged_in": isloggedin})
	})

	// Loging Users
	g.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		logintryst := userstate.CorrectPassword(username, password)

		if logintryst == true {
			userstate.Login(c.Writer, username)
			// c.HTML(http.StatusOK, "index.html", gin.H{"title": "Successful Login"})
			http.Redirect(c.Writer, c.Request, "/", 302)
		} else {

			// c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Invalid credentials provided"})
		}
	})

	//Logout users
	g.GET("/logout", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		userstate.Logout(usercook)
		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	//operator register users
	g.GET("/operator", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "operator.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//operator register users
	g.GET("/kontroler", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "kontroler.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Register visitors POST
	g.POST("/operator", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)

		if isloggedin {

			id := c.PostForm("ID")
			name := c.PostForm("Name")
			nameservice := c.PostForm("NameService")
			age := c.PostForm("Date")
			job := c.PostForm("Number")

			peeps := []*Person{
				{id, usercook, name, nameservice, age, job},
			}

			for _, p := range peeps {
				db.Save(p)
			}

			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Administartort interface
	g.GET("/adminka", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		isadmin := userstate.IsAdmin(usercook)

		var cheked []bool
		if isloggedin {
			listusers, _ := userstate.AllUsernames()
			fmt.Println(isadmin)
			if isadmin {
				for _, i := range listusers {
					fmt.Println(i)
					cheked = append(cheked, userstate.IsAdmin(i))
				}
				fmt.Println(cheked)
				fmt.Println(isadmin)
			}
			c.HTML(http.StatusOK, "adminka.html", gin.H{"listusers": listusers, "is_admin": isadmin, "is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Work page TEST
	// g.GET("/work", func(c *gin.Context) {
	// 	isloggedin := isloggedin(c)
	// 	if isloggedin {
	// 		var person []Person
	// 		err := db.All(&person)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		c.HTML(http.StatusOK, "work.html", gin.H{"person": person, "is_logged_in": isloggedin})
	// 	} else {
	// 		c.AbortWithStatus(http.StatusForbidden)
	// 		fmt.Fprint(c.Writer, "Permission denied!")
	// 	}

	// })

	//Make user as admin POST
	g.GET("/makeadmin/:user", func(c *gin.Context) {
		user := c.Param("user")
		// username := c.PostForm(user)
		userstate.SetAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Delete User from Base POST
	g.GET("/delete/:user", func(c *gin.Context) {
		user := c.Param("user")
		// username := c.PostForm(user)
		userstate.RemoveUser(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Delete Admin status
	g.GET("/adminoff/:user", func(c *gin.Context) {
		user := c.Param("user")
		userstate.IsAdmin(user)
		userstate.RemoveAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	// Start serving the application
	g.Run(":3000")
}
