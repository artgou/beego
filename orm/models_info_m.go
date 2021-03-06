package orm

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

type modelInfo struct {
	pkg       string
	name      string
	fullName  string
	table     string
	model     Modeler
	fields    *fields
	manual    bool
	addrField reflect.Value
}

func newModelInfo(model Modeler) (info *modelInfo) {
	var (
		err error
		fi  *fieldInfo
		sf  reflect.StructField
	)

	info = &modelInfo{}
	info.fields = newFields()

	val := reflect.ValueOf(model)
	ind := reflect.Indirect(val)
	typ := ind.Type()

	info.addrField = ind.Addr()

	info.name = typ.Name()
	info.fullName = typ.PkgPath() + "." + typ.Name()

	for i := 0; i < ind.NumField(); i++ {
		field := ind.Field(i)
		sf = ind.Type().Field(i)
		if field.CanAddr() {
			addr := field.Addr()
			if _, ok := addr.Interface().(*Manager); ok {
				continue
			}
		}
		fi, err = newFieldInfo(info, field, sf)
		if err != nil {
			break
		}
		added := info.fields.Add(fi)
		if added == false {
			err = errors.New(fmt.Sprintf("duplicate column name: %s", fi.column))
			break
		}
		if fi.pk {
			if info.fields.pk != nil {
				err = errors.New(fmt.Sprintf("one model must have one pk field only"))
				break
			} else {
				info.fields.pk.Add(fi)
			}
		}
		if fi.auto {
			info.fields.auto = fi
		}
		fi.fieldIndex = i
		fi.mi = info
	}

	if _, ok := info.fields.pk.Exist(info.fields.auto); info.fields.auto != nil && ok == false {
		err = errors.New(fmt.Sprintf("when auto field exists, you cannot set other pk field"))
		goto end
	}

	if err != nil {
		fmt.Println(fmt.Errorf("field: %s.%s, %s", ind.Type(), sf.Name, err))
		os.Exit(2)
	}

end:
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return
}

func newM2MModelInfo(m1, m2 *modelInfo) (info *modelInfo) {
	info = new(modelInfo)
	info.fields = newFields()
	info.table = m1.table + "_" + m2.table + "_rel"
	info.name = camelString(info.table)
	info.fullName = m1.pkg + "." + info.name

	fa := new(fieldInfo)
	f1 := new(fieldInfo)
	f2 := new(fieldInfo)
	fa.fieldType = TypeBigIntegerField
	fa.auto = true
	fa.pk = true
	fa.dbcol = true

	f1.dbcol = true
	f2.dbcol = true
	f1.fieldType = RelForeignKey
	f2.fieldType = RelForeignKey
	f1.name = camelString(m1.table)
	f2.name = camelString(m2.table)
	f1.fullName = info.fullName + "." + f1.name
	f2.fullName = info.fullName + "." + f2.name
	f1.column = m1.table + "_id"
	f2.column = m2.table + "_id"
	f1.rel = true
	f2.rel = true
	f1.relTable = m1.table
	f2.relTable = m2.table
	f1.relModelInfo = m1
	f2.relModelInfo = m2
	f1.mi = info
	f2.mi = info

	info.fields.Add(fa)
	info.fields.Add(f1)
	info.fields.Add(f2)
	info.fields.pk.Add(fa)
	return
}
