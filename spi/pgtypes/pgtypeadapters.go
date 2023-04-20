package pgtypes

import "github.com/jackc/pglogrepl"

type LSN pglogrepl.LSN

func (lsn LSN) String() string {
	return pglogrepl.LSN(lsn).String()
}

type XLogData struct {
	pglogrepl.XLogData

	LastBegin  LSN
	LastCommit LSN
	Xid        uint32
}

type BeginMessage pglogrepl.BeginMessage

type CommitMessage pglogrepl.CommitMessage

type OriginMessage pglogrepl.OriginMessage

type RelationMessage pglogrepl.RelationMessage

type TypeMessage pglogrepl.TypeMessage

type TruncateMessage pglogrepl.TruncateMessage

type InsertMessage struct {
	*pglogrepl.InsertMessage
	NewValues map[string]any
}

type UpdateMessage struct {
	*pglogrepl.UpdateMessage
	OldValues map[string]any
	NewValues map[string]any
}

type DeleteMessage struct {
	*pglogrepl.DeleteMessage
	OldValues map[string]any
}
