package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxql"
)

func usage() {
	fmt.Println("usage: influx_export_users <path/to/meta.db>")
	os.Exit(1)
}

func main() {
	var (
		all      bool = true
		entusers bool
	)

	flag.BoolVar(&entusers, "entusers", false, "export only users and in enterprise format")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	if entusers {
		all = false
	}

	f, err := os.Open(args[0])
	check(err)
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	check(err)

	data := &meta.Data{}
	err = data.UnmarshalBinary(b)
	check(err)

	if all {
		b, err = json.Marshal(data)
		check(err)
		fmt.Println(string(b))
	} else if entusers {
		usersToEnt(data)
	}
}

func usersToEnt(data *meta.Data) {
	actions := []*UserAction{}
	for i, _ := range data.Users {
		u := &data.Users[i]
		action := &UserAction{
			Action: "create",
			User: &User{
				Name:        u.Name,
				Hash:        u.Hash,
				Permissions: ossPrivsToEntPerms(u),
			},
		}

		actions = append(actions, action)

		b, _ := json.Marshal(action)
		fmt.Println(string(b))
	}
}

func ossPrivsToEntPerms(u *meta.UserInfo) Permissions {
	if u.Admin {
		ps := adminPermissions()
		perms := []string{}
		for _, p := range ps {
			perms = append(perms, p.String())
		}
		return Permissions{"": perms}
	}

	perms := make(Permissions)
	for resource, privilege := range u.Privileges {
		switch privilege {
		case influxql.NoPrivileges:
			perms[resource] = []string{}
		case influxql.ReadPrivilege:
			perms[resource] = []string{ReadDataPermission.String()}
		case influxql.WritePrivilege:
			perms[resource] = []string{WriteDataPermission.String()}
		case influxql.AllPrivileges:
			perms[resource] = []string{ReadDataPermission.String(), WriteDataPermission.String()}
		default:
			perms[resource] = []string{}
		}
	}

	return perms
}

type UserAction struct {
	Action string `json:"action"`
	User   *User  `json:"user"`
}

type User struct {
	Name        string      `json:"name"`
	Password    string      `json:"password"`
	Hash        string      `json:"hash"`
	Permissions Permissions `json:"permissions"`
}

type Permissions map[string][]string

type Permission int

const (
	NoPermissions                   Permission = 0
	ViewAdminPermission             Permission = 1
	ViewChronografPermission        Permission = 2
	CreateDatabasePermission        Permission = 3
	CreateUserAndRolePermission     Permission = 4
	AddRemoveNodePermission         Permission = 5
	DropDatabasePermission          Permission = 6
	DropDataPermission              Permission = 7
	ReadDataPermission              Permission = 8
	WriteDataPermission             Permission = 9
	RebalancePermission             Permission = 10
	ManageShardPermission           Permission = 11
	ManageContinuousQueryPermission Permission = 12
	ManageQueryPermission           Permission = 13
	ManageSubscriptionPermission    Permission = 14
	MonitorPermission               Permission = 15
	CopyShardPermission             Permission = 16
	KapacitorAPIPermission          Permission = 17
	KapacitorConfigAPIPermission    Permission = 18
)

// String returns a string representation of a Permission.
func (p Permission) String() string {
	switch p {
	case NoPermissions:
		return "NoPermissions"
	case ViewAdminPermission:
		return "ViewAdmin"
	case ViewChronografPermission:
		return "ViewChronograf"
	case CreateDatabasePermission:
		return "CreateDatabase"
	case CreateUserAndRolePermission:
		return "CreateUserAndRole"
	case AddRemoveNodePermission:
		return "AddRemoveNode"
	case DropDatabasePermission:
		return "DropDatabase"
	case DropDataPermission:
		return "DropData"
	case ReadDataPermission:
		return "ReadData"
	case WriteDataPermission:
		return "WriteData"
	case RebalancePermission:
		return "Rebalance"
	case ManageShardPermission:
		return "ManageShard"
	case ManageContinuousQueryPermission:
		return "ManageContinuousQuery"
	case ManageQueryPermission:
		return "ManageQuery"
	case ManageSubscriptionPermission:
		return "ManageSubscription"
	case MonitorPermission:
		return "Monitor"
	case CopyShardPermission:
		return "CopyShard"
	case KapacitorAPIPermission:
		return "KapacitorAPI"
	case KapacitorConfigAPIPermission:
		return "KapacitorConfigAPI"
	default:
		return "NoPermissions"
	}
}

func adminPermissions() []Permission {
	return []Permission{
		CreateDatabasePermission,
		CreateUserAndRolePermission,
		DropDataPermission,
		DropDatabasePermission,
		ManageQueryPermission,
		ManageContinuousQueryPermission,
		ManageShardPermission,
		ManageSubscriptionPermission,
		MonitorPermission,
		ReadDataPermission,
		ViewAdminPermission,
		ViewChronografPermission,
		WriteDataPermission,
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
