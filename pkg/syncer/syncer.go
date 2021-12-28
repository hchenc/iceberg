package syncer

type GenerateService interface {
	//add obj to target service
	Add(obj interface{}) (interface{}, error)

	//update obj
	Update(objOld interface{}, objNew interface{}) error

	Delete(name string) error
}

func NewGenerateService(g Generator) GenerateService {
	return generator{g: g}
}

type Generator interface {
	Create(obj interface{}) (interface{}, error)
	//update obj
	Update(objOld interface{}, objNew interface{}) error

	Delete(name string) error

	GetByName(name string) (interface{}, error)

	GetByID(id int) (interface{}, error)

	List(key string) (interface{}, error)
}

type generator struct {
	g Generator
}

func (g generator) Add(obj interface{}) (interface{}, error) {
	return g.g.Create(obj)
}

func (g generator) Update(objOld interface{}, objNew interface{}) error {
	return g.g.Update(objOld, objNew)
}

func (g generator) Delete(name string) error {
	return g.g.Delete(name)
}
