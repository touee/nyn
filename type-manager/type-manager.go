package typemanager

import (
	"reflect"

	"database/sql"
	// sqlite driver
	_ "github.com/mattn/go-sqlite3"

	datapacker "github.com/touee/nyn/data-packer"
)

// TypeManager 管理任务类型
type TypeManager struct {
	packer datapacker.Packer

	db *sql.DB

	nameMap map[string]TypeRecord
}

// TypeRecord 记录了任务类型
type TypeRecord struct {
	Name        string
	ReflectType reflect.Type
}

// OpenTypeManager 打开一个 TypeManager
// 若其并不存在, 创建之
func OpenTypeManager(path string, packer datapacker.Packer) (m *TypeManager, err error) {
	m = new(TypeManager)

	m.packer = packer
	var packFormat, signatureFormat = m.packer.Formats()

	m.db, err = sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}

	var count int
	err = m.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'type_db_meta'`).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count == 0 {
		for _, stmt := range []string{
			`CREATE TABLE type_db_meta (
				meta_key   NOT NULL,
				meta_value NOT NULL
			)`,
			`INSERT INTO type_db_meta (meta_key, meta_value) VALUES ('version', '1')`,
			`CREATE TABLE type_db (
				type_name      TEXT NOT NULL UNIQUE,
				type_signature TEXT NOT NULL
			)`,
		} {
			_, err = m.db.Exec(stmt)
			if err != nil {
				panic(err)
			}
		}

		_, err = m.db.Exec(`INSERT INTO type_db_meta (meta_key, meta_value) VALUES ("pack_format", ?), ("signature_format", ?)`, packFormat, signatureFormat)
		if err != nil {
			panic(err)
		}
	} else {
		var originalPackFormat, originalSignatureFormat string
		err = m.db.QueryRow(`SELECT meta_value FROM type_db_meta WHERE meta_key = ?`, "pack_format").Scan(&originalPackFormat)
		if err != nil {
			panic(err)
		}
		err = m.db.QueryRow(`SELECT meta_value FROM type_db_meta WHERE meta_key = ?`, "signature_format").Scan(&originalSignatureFormat)
		if err != nil {
			panic(err)
		}

		if originalPackFormat != packFormat || originalSignatureFormat != signatureFormat {
			return nil, PackFormatOrSignatureFormatNotMatchError{
				usedPackFormat:          packFormat,
				usedSignatureFormat:     signatureFormat,
				expectedPackFormat:      originalPackFormat,
				expectedSignatureFormat: originalSignatureFormat,
			}
		}
	}

	m.nameMap = make(map[string]TypeRecord)

	return m, nil
}

// GetTypeName 返回任务类型名称, 作为任务类型的唯一标准
func GetTypeName(dummyInstance interface{}) (name string) {
	var (
		t        = reflect.TypeOf(dummyInstance)
		pkgPath  = t.PkgPath()
		typeName = t.Name()
	)
	if pkgPath == "" {
		if typeName != "" {
			return "builtin::" + typeName
		}
		return "::" + t.String()
	}
	return pkgPath + "::" + typeName
}

// Register 在 TypeManager 中注册一个任务类型
func (m *TypeManager) Register(dummyInstance interface{}) (err error) {
	var (
		name      = GetTypeName(dummyInstance)
		signature string
	)
	signature, err = m.packer.Signature(dummyInstance)
	if err != nil {
		return err
	}
	var (
		existedSignature string
	)
	err = m.db.QueryRow(`SELECT type_signature FROM type_db WHERE type_name = ?`, name).Scan(&existedSignature)
	switch {
	case err != nil && err != sql.ErrNoRows: // 遇到错误, 直接返回错误
		panic(err)
	case err == nil && signature != existedSignature:
		return SignatureNotMatchError{name, existedSignature, signature}
	case err == sql.ErrNoRows: // 任务类型并不在数据库中, 创建该任务类型
		_, err = m.db.Exec(`INSERT INTO type_db (type_name, type_signature) VALUES (?, ?)`, name, signature)
		if err != nil {
			panic(err)
		}
		fallthrough
	default: // 数据库中已存在该任务类型的信息
		m.nameMap[name] = TypeRecord{
			Name:        name,
			ReflectType: reflect.TypeOf(dummyInstance),
		}
		return nil
	}
}

// Close 关闭 TypeManager
func (m *TypeManager) Close() (err error) {
	return m.db.Close()
}

// Pack 打包任务
func (m TypeManager) Pack(task interface{}) (packed []byte, err error) {
	var typeName = GetTypeName(task)
	if _, exists := m.nameMap[typeName]; !exists {
		return nil, TypeNotRegisteredError{typeName}
	}
	return m.packer.Pack(task)
}

// Unpack 解包任务
func (m TypeManager) Unpack(typeName string, packed []byte) (task interface{}, err error) {
	var record, exists = m.nameMap[typeName]
	if !exists {
		return nil, TypeNotRegisteredError{typeName}
	}

	var v = reflect.New(record.ReflectType)
	if err = m.packer.Unpack(packed, v.Interface()); err != nil {
		return nil, err
	}
	return reflect.Indirect(v).Interface(), nil
}
