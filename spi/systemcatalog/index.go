package systemcatalog

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/noctarius/timescaledb-event-streamer/internal/supporting"
	"reflect"
	"strings"
	"time"
)

type IndexSortOrder string

const (
	ASC  IndexSortOrder = "ASC"
	DESC IndexSortOrder = "DESC"
)

type IndexNullsOrder string

const (
	NULLS_FIRST IndexNullsOrder = "NULLS FIRST"
	NULLS_LAST  IndexNullsOrder = "NULLS LAST"
)

// Index represents either a (compound) primary key index
// or replica identity index in the database and attached
// to a hypertable (and its chunks)
type Index struct {
	name            string
	columns         []Column
	primaryKey      bool
	replicaIdentity bool
}

func newIndex(name string, column []Column, primaryKey bool, replicaIdentity bool) *Index {
	return &Index{
		name:            name,
		columns:         column,
		primaryKey:      primaryKey,
		replicaIdentity: replicaIdentity,
	}
}

// Name returns the index name
func (i *Index) Name() string {
	return i.name
}

// PrimaryKey returns true if the index represents a
// primary key, otherwise false
func (i *Index) PrimaryKey() bool {
	return i.primaryKey
}

// ReplicaIdentity returns true if the index represents a
// replica identity index, otherwise false
func (i *Index) ReplicaIdentity() bool {
	return i.replicaIdentity
}

// Columns returns an array of colum instances representing
// the columns of the index (in order of definition)
func (i *Index) Columns() []Column {
	return i.columns
}

// AsSqlTuple creates a string to be used as a tuple definition
// in a WHERE-clause:
// (col1, col2, col3, ...)
func (i *Index) AsSqlTuple() string {
	columnList := supporting.Map(i.columns, func(t Column) string {
		return t.Name()
	})
	return fmt.Sprintf("(%s)", strings.Join(columnList, ","))
}

// AsSqlOrderBy creates a string for ORDER BY clauses with all parts
// of the index being ordered in descending direction
func (i *Index) AsSqlOrderBy(desc bool) string {
	order := "ASC"
	if desc {
		order = "DESC"
	}

	columnList := supporting.Map(i.columns, func(t Column) string {
		return fmt.Sprintf("%s %s", t.Name(), order)
	})
	return strings.Join(columnList, ",")
}

// WhereTupleGE creates a WHERE-clause string which selects all values
// greater or equal to the given set of index parameter values
func (i *Index) WhereTupleGE(params map[string]any) (string, bool) {
	return i.whereClause(">=", params)
}

// WhereTupleGT creates a WHERE-clause string which selects all values
// greater than to the given set of index parameter values
func (i *Index) WhereTupleGT(params map[string]any) (string, bool) {
	return i.whereClause(">", params)
}

// WhereTupleLE creates a WHERE-clause string which selects all values
// less or equal to the given set of index parameter values
func (i *Index) WhereTupleLE(params map[string]any) (string, bool) {
	return i.whereClause("<=", params)
}

// WhereTupleLT creates a WHERE-clause string which selects all values
// less than to the given set of index parameter values
func (i *Index) WhereTupleLT(params map[string]any) (string, bool) {
	return i.whereClause("<", params)
}

// WhereTupleEQ creates a WHERE-clause string which selects all values
// equal to the given set of index parameter values
func (i *Index) WhereTupleEQ(params map[string]any) (string, bool) {
	return i.whereClause("=", params)
}

func (i *Index) whereClause(comparison string, params map[string]any) (string, bool) {
	tupleList := i.AsSqlTuple()

	success := true
	comparisonList := supporting.Map(i.columns, func(t Column) string {
		v, present := params[t.name]
		if !present {
			success = false
			return ""
		}
		return param2value(v, t.dataType)
	})

	if !success {
		return "", false
	}

	return fmt.Sprintf("%s %s (%s)", tupleList, comparison, strings.Join(comparisonList, ",")), true
}

func param2value(param any, dataType uint32) string {
	pv := reflect.ValueOf(param)
	pt := pv.Type()

	if pv.Kind() == reflect.Pointer {
		if pv.IsNil() {
			return "NULL"
		}

		pv = reflect.Indirect(pv)
		pt = pv.Type()
	}

	switch mapping[dataType] {
	case FLOAT32, FLOAT64:
		return fmt.Sprintf("%f", pv.Float())
	case INT8, INT16, INT32, INT64:
		return fmt.Sprintf("%d", pv.Int())
	case BOOLEAN:
		if pv.Bool() {
			return "TRUE"
		}
		return "FALSE"
	case STRING:
		val := pv.String()
		if pt.Kind() != reflect.String {
			switch v := pv.Interface().(type) {
			case time.Time:
				val = v.Format(time.RFC3339Nano)
			default:
				panic(errors.Errorf("unhandled string value: %v", pt.String()))
			}
		}
		return fmt.Sprintf("'%s'", sanitizeString(val))

	case BYTES:
		bytes := pv.Interface().([]byte)
		return fmt.Sprintf("bytea '\\x%X'", bytes)
	case ARRAY:
		numOfElements := pv.Len()
		index := 0
		iterator := func() (reflect.Value, bool) {
			if index < numOfElements {
				element := pv.Index(index)
				index++
				return element, true
			}
			return reflect.Zero(pt), false
		}

		elements := supporting.MapWithIterator(iterator, func(element reflect.Value) string {
			return param2value(element.Interface(), 0) // FIXME, right now arrays aren't fully supported
		})
		return fmt.Sprintf("'{%s}'", strings.Join(elements, ","))

	default:
		return reflect.Zero(pt).String()
	}
}

func sanitizeString(val string) string {
	return strings.ReplaceAll(strings.ReplaceAll(val, "'", "\\'"), "\\\\'", "\\'")
}
