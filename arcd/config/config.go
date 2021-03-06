/*
 * Arc - Copyleft of Simone 'evilsocket' Margaritelli.
 * evilsocket at protonmail dot com
 * https://www.evilsocket.net/
 *
 * See LICENSE.
 */
package config

import (
	"encoding/json"
	"errors"
	"github.com/evilsocket/arc/arcd/log"
	"github.com/evilsocket/arc/arcd/utils"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
)

const (
	defAddress         = "127.0.0.1"
	defPort            = 8443
	defMaxReqSize      = int64(512 * 1024)
	defCertificate     = "arcd-tls-cert.pem"
	defKey             = "arcd-tls-key.pem"
	defDatabaseName    = "arc.db"
	defUsername        = "arc"
	defPassword        = "$2a$10$gwnHUhLVV9tgPtZfX4.jDOz6qzGgRHZmtE2YpMr9K1RpIO71YJViO"
	defTokenDuration   = 60
	defSchedulerPeriod = 15
	defBackupsEnabled  = false
	defCompression     = true
	defRateLimit       = 60
)

// SMTP configuration.
type SMTPConfig struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

// PGP configuration.
type PGPConfig struct {
	Enabled bool    `json:"enabled"`
	Keys    KeyPair `json:"keys"`
}

// Reports configuration.
type rpConfig struct {
	Enabled   bool       `json:"enabled"`
	RateLimit int        `json:"rate_limit"`
	Filter    []string   `json:"filter"`
	To        string     `json:"to"`
	PGP       PGPConfig  `json:"pgp"`
	SMTP      SMTPConfig `json:"smtp"`
}

// Scheduler configuration.
type schConfig struct {
	Enabled bool     `json:"enabled"`
	Period  int      `json:"period"`
	Reports rpConfig `json:"reports"`
}

// Backups configuration.
type bkConfig struct {
	Enabled bool   `json:"enabled"`
	Period  int    `json:"period"`
	Folder  string `json:"folder"`
	Run     string `json:"run"`
}

// Arc server configuration.
// swagger:response
type Configuration struct {
	Address       string    `json:"address"`
	Port          int       `json:"port"`
	MaxReqSize    int64     `json:"max_req_size"`
	Certificate   string    `json:"certificate"`
	Key           string    `json:"key"`
	Database      string    `json:"database"`
	Secret        string    `json:"secret"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	TokenDuration int       `json:"token_duration"`
	Compression   bool      `json:"compression"`
	CheckExpired  int       `json:"check_expired"`
	Scheduler     schConfig `json:"scheduler"`
	Backups       bkConfig  `json:"backups"`
}

var Conf = Configuration{
	Address:       defAddress,
	Port:          defPort,
	MaxReqSize:    defMaxReqSize,
	Certificate:   defCertificate,
	Key:           defKey,
	Database:      defDatabaseName,
	Secret:        "",
	Username:      defUsername,
	Password:      defPassword,
	TokenDuration: defTokenDuration,
	Compression:   defCompression,
	Backups: bkConfig{
		Enabled: defBackupsEnabled,
	},
	Scheduler: schConfig{
		Enabled: true,
		Period:  defSchedulerPeriod,
		Reports: rpConfig{
			Enabled:   false,
			RateLimit: defRateLimit,
		},
	},
}

func Load(filename string) error {
	log.Infof("Loading configuration from %s ...", log.Bold(filename))
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &Conf)
	if err != nil {
		return err
	}

	if Conf.Secret == "" {
		return errors.New("HMAC secret not found, please fill the 'secret' configuration field.")
	}

	// fix path
	if Conf.Backups.Folder, err = utils.ExpandPath(Conf.Backups.Folder); err != nil {
		return err
	}

	return nil
}

func (c Configuration) HashPassword(password string, cost int) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func (c Configuration) Auth(username, password string) bool {
	if c.Username != username {
		return false
	}

	if e := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password)); e != nil {
		return false
	}

	return true
}
