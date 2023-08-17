package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent"
	"github.com/sevlyar/go-daemon"
)

func main() {
	var addr string
	var logFile string = carrot.GetEnv("LOG_FILE")
	var runDaemon bool
	var runMigration bool
	var debugLog bool = carrot.GetEnv("DEBUG_LOG") != ""
	var dbDriver string = carrot.GetEnv(carrot.ENV_DB_DRIVER)
	var dsn string = carrot.GetEnv(carrot.ENV_DSN)
	var traceSql bool = carrot.GetEnv("TRACE_SQL") != ""

	var superUserEmail string
	var superUserPassword string

	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&superUserEmail, "superuser", "", "Create an super user with email")
	flag.StringVar(&superUserPassword, "password", "", "Super user password")
	flag.StringVar(&addr, "addr", ":8080", "HTTP Serve address")
	flag.StringVar(&logFile, "log", logFile, "Log output file name, default is os.Stdout")
	flag.BoolVar(&debugLog, "nodebug", debugLog, "Log debug message")
	flag.BoolVar(&runDaemon, "d", false, "Run as daemon")
	flag.BoolVar(&runMigration, "m", false, "Run migration and quit ")
	flag.StringVar(&dbDriver, "db", dbDriver, "DB Driver, sqlite|mysql")
	flag.StringVar(&dsn, "dsn", dsn, "DB DSN")
	flag.BoolVar(&traceSql, "tracesql", traceSql, "Trace sql execution")

	flag.Parse()

	if dsn == "" {
		if _, err := os.Stat(setupDoneFlag); err == nil {
			dsn = "file:restcontent.db"
			if _, err := os.Stat("data"); err == nil {
				dsn = "file:data/restcontent.db"
			}
		} else {
			//
			runSetupMode(addr)
			logFile = carrot.GetEnv("LOG_FILE")
		}
	}

	var lw io.Writer = os.Stdout
	var err error

	if logFile != "" {
		lw, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			log.Printf("open %s fail, %v\n", logFile, err)
		} else {
			log.Default().SetOutput(lw)
		}
	} else {
		logFile = "console"
	}

	if !debugLog {
		carrot.SetLogLevel(carrot.LevelInfo)
	} else {
		carrot.SetLogLevel(carrot.LevelDebug)
	}

	fmt.Println("GitCommit =", GitCommit)
	fmt.Println("BuildTime =", BuildTime)

	fmt.Println("addr      =", addr)
	fmt.Println("logfile   =", logFile)
	fmt.Println("debugLog  =", debugLog)
	fmt.Println("DB Driver =", dbDriver)
	fmt.Println("DSN       =", dsn)

	db, err := carrot.InitDatabase(lw, dbDriver, dsn)
	if err != nil {
		fmt.Println("init database fail", err)
		return
	}

	if traceSql {
		db = db.Debug()
	}

	err = carrot.InitMigrate(db)
	if err != nil {
		panic(err)
	}
	if err := restcontent.Migration(db); err != nil {
		panic(err)
	}
	// Init Database
	if runMigration {
		fmt.Println("migration done")
		return
	}

	if superUserEmail != "" && superUserPassword != "" {
		u, err := carrot.GetUserByEmail(db, superUserEmail)
		if err == nil && u != nil {
			carrot.SetPassword(db, u, superUserPassword)
			carrot.Warning("Update super with new password")
		} else {
			u, err = carrot.CreateUser(db, superUserEmail, superUserPassword)
			if err != nil {
				panic(err)
			}
		}
		u.IsStaff = true
		u.Activated = true
		u.Enabled = true
		u.IsSuperUser = true
		db.Save(u)
		carrot.Warning("Create super user:", superUserEmail)
		return
	}

	r := gin.New()
	m := restcontent.NewManager(db)

	m.GitCommit = GitCommit
	m.BuildTime = BuildTime

	if err = m.Prepare(r, lw); err != nil {
		log.Panic("prepare restcontent fail", err)
		return
	}
	if addr[0] == ':' {
		addr = "localhost" + addr
	}
	fmt.Println("restcontent server is running on ", "http://"+addr)
	if runDaemon {
		cntxt := &daemon.Context{
			WorkDir: ".",
		}
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal("Unable to run: ", err)
		}
		if d != nil {
			return
		}
		defer cntxt.Release()
		r.Run(addr)
	} else {
		r.Run(addr)
	}
}
