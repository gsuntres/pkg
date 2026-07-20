package mongo

type Op string

const (
	OpUpdate Op = "update"
	OpDelete Op = "delete"
	OpInsert Op = "insert"
	OpArchive Op = "archive"
	OpRestore Op = "restore"
)

type AuditResult struct {
	Op Op `bson:"op" json:"op"`
	Entity string  `bson:"entity,omitempty" json:"entity,omitempty"`
	Data any  `bson:"data,omitempty" json:"data,omitempty",`
	Before any  `bson:"before,omitempty" json:"before,omitempty"`
	After any  `bson:"after,omitempty" json:"after,omitempty"`
	Context any  `bson:"context,omitempty" json:"context,omitempty"`
}

type OnAudit func(audit *AuditResult) error