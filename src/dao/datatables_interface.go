package dao

type EditDataTables interface {
	Insert() (error)
	Remove() (error)
	GetfileId()(string, error)
	GetOneById()(interface{})
	Update()(error)
	GetAll()([]interface{}, error)
}