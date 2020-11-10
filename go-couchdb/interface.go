package couchdb

/* This defines the interfaces we need to conform to for proper compatibility */

type clientIface interface {
	AllDBs() ([]string, error)
	CreateDB(string) (*DB, error)
	DB(string) *DB
	DBUpdates(Options) (*DBUpdatesFeed, error)
	DeleteDB(string) error
	EnsureDB(string) (*DB, error)
	Ping() error
	SetAuth(Auth)
	URL() string
}

type dbIface interface {
	AllDocs(interface{}, Options) error
	Attachment(string, string, string) (*Attachment, error)
	AttachmentMeta(string, string, string) (*Attachment, error)
	Changes(Options) (*ChangesFeed, error)
	Delete(string, string) (string, error)
	DeleteAttachment(string, string, string) (string, error)
	Get(string, interface{}, Options) error
	Name() string
	Put(string, interface{}, string) (string, error)
	PutAttachment(string, *Attachment, string) (string, error)
	PutSecurity(*Security) error
	Rev(string) (string, error)
	Security() (*Security, error)
	View(string, string, interface{}, Options) error
}

type updatesIface interface {
	Close() error
	Err() error
	Next() bool
}

type changesIface interface {
	ChangesRevs() []string
	Close() error
	Err() error
	Next() bool
}
