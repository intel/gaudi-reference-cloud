// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/configmgr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/grpcserver"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/handlers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/ovnclient"

	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	/*
		"github.com/ovn-org/libovsdb/ovsdb"
		libovsdbclient "github.com/ovn-org/libovsdb/client"
		libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
		libovsdbutil "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/util"
		"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/config"
		"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
		"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/sbdb"
		"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/types"

	*/)

var (
	nbClient libovsdbclient.Client
	dbClient *sql.DB
	Logger   logr.Logger
	cfg      configmgr.Config
	ipaddr   *string
	ovnport  *string
	srvport  *string
	dbipaddr *string
	dbport   *string
	cfgfile  *string
)

func controllerInit(ctx context.Context) {
	initFlags()
	initLogger(ctx)

	// getConfig from CLI flag args or from cfg file
	getConfig(ctx)
	logger := Logger.WithName("ControllerInit")

	// Connect to the database

	if (manageddb.Config{}) != cfg.Database {
		logger.Info("Using manageddb based method for SQL DB connection")
		logger.Info(fmt.Sprint("URL for manageddb is: ", cfg.Database.URL))
		logger.Info(fmt.Sprint("Reading config from file: ", *cfgfile))
		managedDB, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "Failed to init Db ")
			return
		}

		dbClient, err = managedDB.Open(ctx)
		if err != nil {
			logger.Error(err, "Failed to open Db ")
			return
		}
	} else {
		logger.Info("Using conventional method for SQL DB connection")
		connectDB()
	}

	// try to initialize the tables
	err := createSDNVNTables(dbClient, "/sdn-vn-tables-creation.sql")
	if err != nil {
		panic(err)
	}

	// Start the OVN Client
	startOvnClient(ctx)

	if err := handlers.CreateGateways(dbClient, nbClient, cfg.GatewaysCfg); err != nil {
		panic(err)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*100))
	defer cancel()

	controllerInit(ctx)

	logger := Logger.WithName("main")
	logger.Info("Starting Server...")
	grpcserver.Server(ctx, &cfg.GrpcServerCfg.Srvport, nbClient, dbClient, &cfg.SecurityCfg.GrpcSsl)
	logger.Info("Done!")
}

// Initialize the command line flags
func initFlags() {
	// Check if sdncontroller has cmdline args: -ip=<hostname> -ovn-port=<port>, etc.
	// If user does not set flags for ip, ovn-port, etc. we will use a blank ("") for now
	// cfgfile will use default value of "config.yaml"
	ipaddr = flag.String("ip", "", "IP address or a hostname of OVN Central")
	ovnport = flag.String("ovn-port", "", "Port Number of OVN Central")
	srvport = flag.String("api-server-port", "", "API Server Port Number")
	dbipaddr = flag.String("db-ip", "", "IP address or a hostname of PostgreSQL")
	dbport = flag.String("db-port", "", "Port Number of PostgreSQL")
	cfgfile = flag.String("config", "config.yaml", "Config file")

	// BindFlags will parse the given flagset for zap option flags and set the
	// log options accordingly. BindFlags needs to be done before flag Parse
	log.BindFlags()
	flag.Parse() // needs to be done before SetDefaultLogger

	// Exit if there are non flag args (NArg).
	if flag.NArg() > 0 {
		fmt.Println("Unexpected arguments:", flag.Args())
		flag.Usage()
		fmt.Printf("\nExamples:\n\t%s --ip=localhost\n", os.Args[0])
		fmt.Printf("\t%s -ip=10.10.10.5 -ovn-port=6641 -api-server-port=50051\n\n", os.Args[0])
		os.Exit(0)
	}
}

// Initialize the logger
func initLogger(ctx context.Context) {
	log.SetDefaultLogger()

	// Logger for the main package
	Logger = log.FromContext(ctx).WithName("main")
	logger := Logger.WithName("initLogger")

	// Main pkg can see global variables other sub packages but not vice versa.
	// So initialize the global Logger variables of the other packages
	handlers.Logger = log.FromContext(ctx).WithName("handlers")
	ovnclient.Logger = log.FromContext(ctx).WithName("ovnclient")
	grpcserver.Logger = log.FromContext(ctx).WithName("grpcserver")

	v := logger.GetV()
	logger.Info(fmt.Sprintf("Logger Verbose level %d", v))
}

func initConfig() {
	// Initialize cfg struct with default values - it will get overwritten
	// by the values read from the command line flags or the config file
	cfg.OvnCentralCfg.Ipaddr = "localhost"
	cfg.OvnCentralCfg.Ovnport = "6641"
	cfg.GrpcServerCfg.Srvport = "50051"
	cfg.DbCfg.DbAvailable = ""
	cfg.DbCfg.Ipaddr = "localhost"
	cfg.DbCfg.Dbport = "5432"
	cfg.DbCfg.Dbname = "sdncontroller"
	cfg.DbCfg.Dbuser = "sdncontroller"
	cfg.DbCfg.Dbpasswd = ""

	for _, sslCfg := range []*configmgr.SslConfig{
		&cfg.SecurityCfg.SqlSsl,
		&cfg.SecurityCfg.OvnSsl,
		&cfg.SecurityCfg.GrpcSsl} {
		sslCfg.Enabled = ""
		sslCfg.Verify = ""
		sslCfg.Ca = ""
		sslCfg.Cert = ""
		sslCfg.Key = ""
	}

}

// getConfig from CLI flag args or from cfg file
func getConfig(ctx context.Context) {
	initConfig()
	logger := Logger.WithName("getConfig")
	logger.V(1).Info(fmt.Sprintf("Initial cfg struct is: %+v", cfg))

	logger.V(1).Info(fmt.Sprint("Cmd line args: ipaddr:", *ipaddr, ", ovn port:", *ovnport,
		", api server port:", *srvport))

	// We want to read cfg even if cmd args are present as there may be other cfg params
	// Load configuration from file.
	logger.Info(fmt.Sprint("Reading config from file: ", *cfgfile))
	if err := conf.LoadConfigFile(ctx, *cfgfile, &cfg); err != nil {
		logger.Error(err, "Failed to load configuration")
		panic(err)
	}

	// Hide Dbpasswd for security reasons before displaying the config
	Dbpasswd := cfg.DbCfg.Dbpasswd
	cfg.DbCfg.Dbpasswd = "*********"
	logger.V(1).Info(fmt.Sprintf("Updated cfg struct after reading cfg file is: %+v", cfg))
	logger.V(1).Info(fmt.Sprintf("Gateways config is: %+v", cfg.GatewaysCfg))

	// If the command line fields have non-blank user provided values, update the
	// cfg struct with those values. Otherwise the cfg struct will have the default
	// values it was initialized with
	if *ipaddr != "" {
		cfg.OvnCentralCfg.Ipaddr = *ipaddr
	}
	if *ovnport != "" {
		cfg.OvnCentralCfg.Ovnport = *ovnport
	}
	if *srvport != "" {
		cfg.GrpcServerCfg.Srvport = *srvport
	}
	if *dbipaddr != "" {
		cfg.DbCfg.Ipaddr = *dbipaddr
	}
	if *dbport != "" {
		cfg.DbCfg.Dbport = *dbport
	}
	logger.V(1).Info(fmt.Sprintf("Final cfg struct is: %+v", cfg))

	// Restore the password
	cfg.DbCfg.Dbpasswd = Dbpasswd
}

// Start the OVN Client
func startOvnClient(ctx context.Context) {
	var err error
	logger := Logger.WithName("startOvnClient")
	s := fmt.Sprint(cfg.OvnCentralCfg.Ipaddr, ":", cfg.OvnCentralCfg.Ovnport)
	logger.Info(fmt.Sprint("Starting OvnClient: ", s))

	nbClient, err = ovnclient.NewNBClient(s, ctx, &cfg.SecurityCfg.OvnSsl)
	if err != nil {
		panic(err)
	}
	logger.Info("OvnClient established")
}

// connect to the database
func connectDB() {
	var err error
	logger := Logger.WithName("connectDB")
	var dbAvailable bool
	dbAvailable, err = strconv.ParseBool(cfg.DbCfg.DbAvailable)
	if err != nil {
		logger.Error(err, "Cannot parse db available cfg.\n")
		return
	}
	if !dbAvailable {
		// Connect to the DB only if available
		logger.Info("Config file says database is not available. " +
			"If you need DB, set 'Database available' as yes in the cfg file")
		return
	}

	dbConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s",
		cfg.DbCfg.Ipaddr, cfg.DbCfg.Dbport,
		cfg.DbCfg.Dbuser, cfg.DbCfg.Dbname)
	// Apply security options when specified
	var sslEnabled bool
	sslEnabled, err = strconv.ParseBool(cfg.SecurityCfg.SqlSsl.Enabled)
	if err != nil {
		logger.Error(err, "Cannot parse SSL enabled cfg for database.\n")
		return
	}
	if !sslEnabled {
		dbConnStr += " sslmode=disable"
	} else {
		dbConnStr += fmt.Sprintf(
			" sslrootcert=%s sslkey=%s sslcert=%s",
			cfg.SecurityCfg.SqlSsl.Ca,
			cfg.SecurityCfg.SqlSsl.Key,
			cfg.SecurityCfg.SqlSsl.Cert,
		)
		var sslVerify bool
		sslVerify, err = strconv.ParseBool(cfg.SecurityCfg.SqlSsl.Verify)
		if err != nil {
			logger.Error(err, "Cannot parse SSL verify cfg for database.\n")
			return
		}
		if !sslVerify {
			dbConnStr += "sslmode=require"
		} else {
			dbConnStr += "sslmode=verify-full"
		}
	}
	if cfg.DbCfg.Dbpasswd != "" {
		/*
			passBytes, err := os.ReadFile(cfg.DbCfg.Dbpasswd)
			if err != nil {
				panic(err)
			}
		*/
		passBytes := cfg.DbCfg.Dbpasswd
		dbConnStr = dbConnStr + " password=" + string(passBytes)
	}

	dbClient, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		panic(err)
	}
	if err = dbClient.Ping(); err != nil {
		if err := dbClient.Close(); err != nil {
			Logger.Error(err, "Failed to close db connection")
		}
		panic(err)
	}
	logger.Info("PostgreSQL connected")
}

// This function is just a workaround to run the table creation script, and this is a very simple approach for getting the statements,
// it need to be improved if we have complicate SQL, or remove this logic to have other approach to init the tables.
func createSDNVNTables(db *sql.DB, filePath string) error {
	logger := Logger.WithName("createSDNVNTables")

	// Read the SQL file from disk
	sqlBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Split the file into separate statements if necessary.
	statements := strings.Split(sqlStr, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		// Skip empty statements
		if stmt == "" {
			continue
		}

		logger.Info(fmt.Sprintf("executing SQL statement: %s", stmt))
		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed executing statement [%s]: %v", stmt, err)
		}
	}
	logger.Info("tables initialization done")
	return nil
}
