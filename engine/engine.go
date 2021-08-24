package engine

import "container/list"
import "fmt"
//import "log"
import "reflect"

type Variable string

type Operator int

const ( //enumeration
	EQ Operator = iota
	GE
	GT
	LE
	LT
	NE //not equal to
)

func (op Operator) String() string {

	switch op {
		case EQ:
			return "EQ"
		case GE:
			return "GE"
		case GT:
			return "GT"
		case LE:
			return "LE"
		case LT:
			return "LT"
		case NE:
			return "NE"
		default:
			return ""
	}
}

type Fact struct {
	ObjectId  string
	Attribute string
	Value     interface{} //scalars only (in this version)
}

func (fact Fact) String() string {
//this is primarily for debugging
	reflectedValue := reflect.ValueOf(fact.Value)

	switch reflectedValue.Kind() {
	case reflect.String:
		return fmt.Sprintf("O %s A %s V %s",fact.ObjectId,fact.Attribute,reflectedValue.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("O %s A %s V %d",fact.ObjectId,fact.Attribute,reflectedValue.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("O %s A %s V %f",fact.ObjectId,fact.Attribute,reflectedValue.Float())
	default:
		return fmt.Sprintf("O %s A %s V.Kind %s V.Type %s",fact.ObjectId,fact.Attribute, reflectedValue.Kind().String(), reflectedValue.Type().Name())
	}
}

type Condition struct {
	NotExists  bool        //negation of existential quantification
	ObjectId   interface{} //string or Variable
	Attribute  string
	Comparator Operator
	Value      interface{}
}

type Rule struct {
	Id  string
	LHS []Condition
	RHS []Inference
}

type Inference struct {
	ObjectId  interface{} //string or Variable
	Attribute string
	Value     interface{}
}

type Engine struct {

	agenda list.List //used as FIFO queue of *Fact
	alphaNetwork map[string][]*alphaNode //keyed by attribute
	nullFact Fact
	productions []*pNode
}

func (engine Engine) GetInferences(objectId string, attribute string) ([]Fact, error) {
	
	var list []Fact
	for _, p := range engine.productions {
		for _, t := range p.tokens {
			for _, f := range t.outgoing {
				if f != nil && (objectId == "" || f.ObjectId == objectId) && (attribute == "" || f.Attribute == attribute) {
					list = append(list, *f)
				}
			}
		}
	}
	return list, nil
}

func (engine *Engine) Assert(fct Fact) (err error) {	

	engine.pushAgenda(&fct)
	err = engine.turn()
	if err != nil {
		return err
	}
	return nil
}

func (engine *Engine) Retract(fct Fact) (err error) {

	f, err := engine.find(fct)
	if err != nil {
		return err
	}

	err = engine.retract(f)
	if err != nil {
		return err
	}

	return nil
}

func (engine *Engine) Define(r Rule) (err error) {

	var newAlphaNode *alphaNode
	var newBetaNode *betaNode
	var newPNode *pNode

	/*if this is the first time Define has been run for this engine
	  then the alpha network must be initialized */
	if engine.alphaNetwork == nil {
		engine.nullFact = Fact{} //represents an absence of facts
		engine.alphaNetwork = make(map[string][]*alphaNode,5) //keyed by attribute
	}

	//create the p-node, but do not add inferences, yet
	newPNode = &pNode{}
	newPNode.parentEngine = engine
	newPNode.ruleId = r.Id
	newPNode.testNetwork = make(map[Variable][]betaTest,5)
	engine.productions = append(engine.productions, newPNode)

	//iterate over the conditions
	for i, condition := range r.LHS {
		newAlphaNode = nil
		//if this attribute hasn't been seen before, add it
		_, ok := engine.alphaNetwork[condition.Attribute]
		if !ok {
			engine.alphaNetwork[condition.Attribute] = make([]*alphaNode,0)
		}
		condObjIdType := reflect.TypeOf(condition.ObjectId).Name()
		condValueType := reflect.TypeOf(condition.Value).Name()
		//error condition check:
		if condValueType == "Variable" && condition.Comparator != EQ {
			return fmt.Errorf("Value variable cannot be used with %s",condition.Comparator.String())
		}
		tempNode := alphaNode{}
		tempNode.parentEngine = engine
		tempNode.attributeName = condition.Attribute
		if condObjIdType != "Variable" {
			tempNode.objConstraint = condition.ObjectId.(string)
		}
		tempNode.comparator = condition.Comparator
		if condValueType != "Variable" {
			tempNode.compareTo = condition.Value
		}
		//if an alpha node already exists with these features, re-use it
		for j, compareNode := range engine.alphaNetwork[condition.Attribute] {
			compareToMatched, err := match(tempNode.compareTo,EQ,compareNode.compareTo)
			if err != nil {
				return err
			}
			if tempNode.objConstraint == compareNode.objConstraint && 
			   tempNode.comparator == compareNode.comparator && 
			   compareToMatched {
				newAlphaNode = engine.alphaNetwork[condition.Attribute][j]
				break
			}
		}
		//otherwise, add a new one
		if newAlphaNode == nil {
			newAlphaNode = &tempNode
			engine.alphaNetwork[condition.Attribute] = append(engine.alphaNetwork[condition.Attribute], newAlphaNode)
		}

		/*create an empty beta node
		  (the alpha nodes are housed in a map, but the
		   beta nodes just float out there in memory space) */
		newBetaNode = &betaNode{}

		//start populating it
		newAlphaNode.betaNodes = append(newAlphaNode.betaNodes, newBetaNode)
		newBetaNode.index = i
		newBetaNode.parentNode = newAlphaNode
		newBetaNode.product = newPNode
		newPNode.betaNodes = append(newPNode.betaNodes, newBetaNode)

		if condition.NotExists {
			newBetaNode.alphaNot = true
		}

		//process the variables (if any) into the p-node's test network
		if condObjIdType == "Variable" {
			//if this variable hasn't been seen before, add it
			_, ok := newPNode.testNetwork[condition.ObjectId.(Variable)]
			if !ok {
				 newPNode.testNetwork[condition.ObjectId.(Variable)] = make([]betaTest,0)
			}
			tmp := betaTest {
					tokenIndex: i,
					objectElseValue: true,
			}
			//the beta node will remember where it is
			newBetaNode.objectVariable = condition.ObjectId.(Variable)
			newBetaNode.objectIndex = len(newPNode.testNetwork[condition.ObjectId.(Variable)])
			newPNode.testNetwork[condition.ObjectId.(Variable)] = append(newPNode.testNetwork[condition.ObjectId.(Variable)], tmp)
		}
		if condValueType == "Variable" {
			//if this variable hasn't been seen before, add it
			_, ok := newPNode.testNetwork[condition.Value.(Variable)]
			if !ok {
				 newPNode.testNetwork[condition.Value.(Variable)] = make([]betaTest,0)
			}
			tmp := betaTest {
					tokenIndex: i,
					objectElseValue: false,
			}
			newBetaNode.valueVariable = condition.Value.(Variable)
			newBetaNode.valueIndex = len(newPNode.testNetwork[condition.Value.(Variable)])
			newPNode.testNetwork[condition.Value.(Variable)] = append(newPNode.testNetwork[condition.Value.(Variable)], tmp)
		}
	}

	//add inferences to p-node
	newPNode.inferences = r.RHS

	return nil
}

/*******************************************************************/
/* the public interface is above; everything below is non-exported */
/*******************************************************************/

func (engine *Engine) find(fct Fact) (*Fact, error) {

	for _, node := range engine.alphaNetwork[fct.Attribute] {
		for _, fptr := range node.facts {
			matched, err := match(fptr.Value,EQ,fct.Value)
			if err != nil {
				return nil, err
			}
			if fptr.ObjectId == fct.ObjectId &&
			   fptr.Attribute == fct.Attribute &&
			   matched {
				return fptr, nil
			}
		}
	}
	return nil, nil
}

func (engine *Engine) retract(f *Fact) (err error) {

	if f == nil {
		return fmt.Errorf("Cannot retract nil")
	}

	for i, node := range engine.alphaNetwork[f.Attribute] {
		for j, v := range node.facts {
			if v == f {
				engine.alphaNetwork[f.Attribute][i].removeFact(j)
			}
		}
	}

	return nil
}

func (engine *Engine) printAlphaNetwork() {
	//this is for debugging
	for k, nodeList := range engine.alphaNetwork {
		fmt.Printf("NodeList for %s\n", k)
		for i, node := range nodeList {
			fmt.Printf("\tNode: %d\n", i)
			fmt.Printf("\tattributeName: %s\n", node.attributeName)
			fmt.Printf("\tobjConstraint: %s\n", node.objConstraint)
			fmt.Printf("\tcomparator: %s\n",node.comparator.String())
			fmt.Printf("\tcompareTo: %v\n", node.compareTo)
			fmt.Printf("\tNo.Facts: %d\n", len(node.facts))
			fmt.Printf("\tNo.Betas: %d\n", len(node.betaNodes))
			for _, f := range node.facts {
				fmt.Printf("\t\t%s\t%s\t%v\n", f.ObjectId, f.Attribute, f.Value)
			}
		}
	}
}

func (engine *Engine) printBetaNetwork() {
	//this is for debugging
	for _, p := range engine.productions{
		fmt.Printf("%s\n",p.ruleId)
		for _, b := range p.betaNodes {
			fmt.Printf("\tBeta Node %d, negation %t\n",b.index,b.alphaNot)
		}
		/* unfinished */
	}
}

func (engine Engine) printTokens() error {
	//this is for debugging	
	for _, p := range engine.productions {
		fmt.Println(p.ruleId)
		for i, t := range p.tokens {
			fmt.Printf("token %d\n",i)
			for _, fct := range t.incoming {
				if fct != nil {
					fmt.Printf("O: %s, A: %s, V: %v\n",fct.ObjectId,fct.Attribute,fct.Value)
				} else {
					fmt.Println("nil")
				}
			}
			fmt.Println("Inferences:")
			for _, f := range t.outgoing {
				if f != nil {
					fmt.Printf("O: %s, A: %s, V: %v\n",f.ObjectId,f.Attribute,f.Value)
				} else {
					fmt.Println("nil")
				}
			}
		}
	}
	return nil
}

func (engine *Engine) pushAgenda(f *Fact) (err error) {

	_ = engine.agenda.PushFront(f)

	return nil
}

func (engine *Engine) popAgenda() (f *Fact, err error) {

	pop := engine.agenda.Back()

	if pop == nil {
		return nil, nil
	} else {
		f = engine.agenda.Remove(pop).(*Fact)
		return f, nil
	}
}

//this is the main action loop: 
//the engine "turns" until the agenda is empty
func (engine *Engine) turn() error {

	var duplicateAssertion bool

	for {
		duplicateAssertion = false
		f, err := engine.popAgenda()
		if err != nil {
			return err
		}
		if f == nil {
			break
		}
		//propagate f into the alpha network
		alphaList, ok := engine.alphaNetwork[f.Attribute]
		if ok {
			for i, aNode := range alphaList {
				if len(aNode.objConstraint) > 0 && f.ObjectId != aNode.objConstraint {
					continue
				}
				if aNode.compareTo != nil {
					matched, err := match(f.Value, aNode.comparator, aNode.compareTo)
					if err != nil {
						return err
					}
					if !matched {
						continue
					}
				}
				//check for duplication (this is inefficient)
				for _, existing := range alphaList[i].facts {
					valuesMatch, err := match(f.Value,EQ,existing.Value)
					if err != nil {
						return err
					}
					if f.ObjectId == existing.ObjectId && valuesMatch {
						duplicateAssertion = true
						break	
					}
				}
				if duplicateAssertion {
					break
				}
				//f hasn't been disqualified, so add it to the alpha node
				alphaList[i].facts = append(alphaList[i].facts, f)
				//right activate all of the beta nodes
				for _, bNode := range aNode.betaNodes {
					err = bNode.rightActivate(f)
					if err != nil {
						return err
					}
				}
			}
		} //else f is an irrelevant or duplicate fact
	}
	return nil
}

type alphaNode struct {

	parentEngine *Engine

	attributeName string
	objConstraint string //object id equals

	comparator Operator
	compareTo  interface{} //scalars only (in this version)

	facts []*Fact
	betaNodes []*betaNode
}

func (node *alphaNode) removeFact(i int) error {

	if node == nil {
		return fmt.Errorf("removeFact: node is nil\n")
	}

	if i >= len(node.facts) {
		return fmt.Errorf("removeFact: index is out of range\n")
	}

	node.facts[i] = node.facts[len(node.facts)-1]
	node.facts[len(node.facts)-1] = nil
	node.facts = node.facts[:len(node.facts)-1]

	for _, bNode := range node.betaNodes {
		for _, t := range bNode.product.tokens {
			if len(node.facts) == 0 && bNode.alphaNot {//existential negation
				t.incoming[bNode.index] = &node.parentEngine.nullFact
				err := bNode.leftActivate(t)
				if err != nil {
					return err
				}
			} else {
				t.damage(bNode.index)
			}
		}
		if len(node.facts) == 0 && bNode.alphaNot && len(bNode.product.tokens) == 0 {
			t, err := bNode.product.addToken(nil,0) //existential negation
			if err != nil {
				return err
			}
			err = bNode.leftActivate(t)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (node alphaNode) String() string {
//this is primarily for debugging
	reflectedValue := reflect.ValueOf(node.compareTo)

	switch reflectedValue.Kind() {
	case reflect.String:
		return fmt.Sprintf("A %s O %s %s V %s F %d B %d",node.attributeName,node.objConstraint,node.comparator.String(),reflectedValue.String(),len(node.facts),len(node.betaNodes))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("A %s O %s %s V %d F %d B %d",node.attributeName,node.objConstraint,node.comparator.String(),reflectedValue.Int(),len(node.facts),len(node.betaNodes))
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("A %s O %s %s V %f F %d B %d",node.attributeName,node.objConstraint,node.comparator.String(),reflectedValue.Float(),len(node.facts),len(node.betaNodes))
	default:
		return fmt.Sprintf("A %s O %s V.Kind %s F %d B %d",node.attributeName,node.objConstraint, reflectedValue.Kind().String(), len(node.facts),len(node.betaNodes))
	}
}

type betaNode struct {

	index int //condition number

	alphaNot bool //existential negation

	objectVariable Variable
	objectIndex int

	valueVariable Variable
	valueIndex int

	//these must not be nil:
	parentNode *alphaNode
	product *pNode
}

//rightActivate handles the arrival of a single new fact
func (node *betaNode) rightActivate(newFact *Fact) (err error) {

	if node == nil {
		return fmt.Errorf("rightActivate: node is nil\n")
	}

	if newFact == nil {
		return fmt.Errorf("rightActivate: received nil Fact pointer\n")
	}

	var tokenFound bool

	for i, _ := range node.product.tokens {
		success, err := node.product.tokens[i].inject(newFact, node)
		if err != nil {
			return err
		}
		if success {
			tokenFound = true
			node.leftActivate(node.product.tokens[i])
		}
	}

	if tokenFound {
		return nil
	}

	//restrict new token creation to right activation
	//new token creation should only occur here
	tok, err := node.product.addToken(newFact, node.index)
	if err != nil {
		return err
	}

	err = node.leftActivate(tok)
	if err != nil {
		return err
	}

	return nil
}

//leftActivate tries to fill out the rest of a single token with multiple facts, 
//skipping the location that belongs to this node because it is assumed to have 
//just been populated prior to this method being called
func (node *betaNode) leftActivate(tok *token) (err error) {

	if node == nil {
		return fmt.Errorf("leftActivate: node is nil\n")
	}

	if tok == nil {
		return fmt.Errorf("leftActivate: received nil token\n")
	}

	for i := 0; i < len(tok.incoming); i++ {
		//this statement is unnecessary, but helps make the
		if i == node.index {//concept of left activation clear
			continue
		}
		if tok.incoming[i] == nil {
			for _, f := range node.product.betaNodes[i].parentNode.facts {
				success, err := tok.inject(f, node.product.betaNodes[i])
				if err  != nil {
					return err
				}
				if success {
					break
				}
			}
		}
	}

	err = node.product.activate(tok)
	if err != nil {
		return err
	}

	return nil
}

//inject tries to add a single fact to a single token
func (tok *token) inject(newFact *Fact, node *betaNode) (bool, error) {

	if newFact == nil {
		return false, fmt.Errorf("inject: received nil Fact pointer")
	}

	if tok == nil {
		return false, fmt.Errorf("inject: received nil token pointer")
	}

	if node == nil {
		return false, fmt.Errorf("inject: node is nil")
	}

	existing := tok.incoming[node.index]

	if node.alphaNot {
		if existing == &node.product.parentEngine.nullFact {
			//existential negation is backwards
			tok.damage(node.index)
			return true, nil
		} else if existing != nil {
			return false, fmt.Errorf("inject: invalid token")
		} else {
			return true, nil
		}
	}

	if existing != nil {

		valuesMatch, err := match(newFact.Value,EQ,existing.Value)

		if err != nil {
			return false, err
		}

		if newFact.ObjectId == existing.ObjectId && newFact.Attribute == existing.Attribute && valuesMatch {
			return true, nil
		}
		return false, nil
	}

	//if the beta node has an object variable, run through the test network
	if node.objectVariable != "" {
		for key, slc := range node.product.testNetwork {
			for _, tst := range slc {
				if tok.incoming[tst.tokenIndex] == nil {
					continue
				}
				if node.objectVariable == key {//EQ
					if tst.objectElseValue {
						if newFact.ObjectId != tok.incoming[tst.tokenIndex].ObjectId {
							return false, nil
						}
					} else {
						val, ok := tok.incoming[tst.tokenIndex].Value.(string)
						if !ok || newFact.ObjectId != val {
							return false, nil
						}
					}
				} else {//NE
					if tst.objectElseValue {
						if newFact.ObjectId == tok.incoming[tst.tokenIndex].ObjectId {
							return false, nil
						}
					} else {
						val, ok := tok.incoming[tst.tokenIndex].Value.(string)
						if ok && newFact.ObjectId == val {
							return false, nil
						}
					}
				}
			}
		}
	}

	//if the beta node has a value variable, run through the test network
	if node.valueVariable != "" {
		for key, slc := range node.product.testNetwork {
			for _, tst := range slc {
				if tok.incoming[tst.tokenIndex] == nil {
					continue
				}
				if node.valueVariable == key {//EQ
					if tst.objectElseValue {
						val, ok := newFact.Value.(string)
						if !ok || val != tok.incoming[tst.tokenIndex].ObjectId {
							return false, nil
						}
					} else {
						notMatched, err := match(newFact.Value, NE, tok.incoming[tst.tokenIndex].Value)
						if err != nil {
							return false, err
						}
						if notMatched {
							return false, nil
						}
					}
				} else {//NE
					if tst.objectElseValue {
						val, ok := newFact.Value.(string)
						if ok && val == tok.incoming[tst.tokenIndex].ObjectId {
							return false, nil
						}
					} else {
						matched, err := match(newFact.Value, EQ, tok.incoming[tst.tokenIndex].Value)
						if err != nil {
							return false, err
						}
						if matched {
							return false, nil
						}
					}
				}
			}
		}
	}

	//the fact has successfully run the gauntlet, so add it
	tok.incoming[node.index] = newFact
	//newFact.tokens = append(newFact.tokens,tok)
	return true, nil
}

type token struct {

	containedBy *pNode
	incoming []*Fact
	outgoing []*Fact
}

func (t token) print() {

	fmt.Printf("IN |")
	for _, fct := range t.incoming {
		if fct != nil {
			fmt.Printf("O: %s, A: %s, V: %v|",fct.ObjectId,fct.Attribute,fct.Value)
		} else {
			fmt.Printf("nil|")
		}
	}
	fmt.Printf("\nOUT|")
	for _, f := range t.outgoing {
		if f != nil {
			fmt.Printf("O: %s, A: %s, V: %v|",f.ObjectId,f.Attribute,f.Value)
		} else {
			fmt.Printf("nil|")
		}
	}
	fmt.Printf("\n")
	return
}

func (t *token) damage(i int) error {

	fct := t.incoming[i]

	if fct == nil {
		return fmt.Errorf("damage: token is not nil at index")
	}

	//take out the requested location
	t.incoming[i] = nil

	//now retract all inferences
	for j, f := range t.outgoing {
		err := t.containedBy.parentEngine.retract(f)
		if err != nil {
			return err
		}
		t.outgoing[j] = nil
	}
	//check if token is now empty
	for _, f := range t.incoming {
		if f != nil {
			return nil
		}
	}
	//if so, self-destruct
	err := t.containedBy.removeToken(t)
	if err != nil {
		return err
	}
	return nil
}

type betaTest struct {
	tokenIndex int
	objectElseValue bool
}

type pNode struct {

	ruleId string

	parentEngine *Engine

	betaNodes  []*betaNode //ordered
	tokens     []*token
	inferences []Inference

	testNetwork map[Variable][]betaTest
}

func (node *pNode) addToken(f *Fact, i int) (t *token, err error) {

	if node.betaNodes[i].alphaNot && f != nil {
		//existential negation is backwards
		return nil, nil
	}

	t = &token{}
	node.tokens = append(node.tokens, t)
	t.containedBy = node
	t.incoming = make([]*Fact,len(node.betaNodes))
	t.outgoing = make([]*Fact,len(node.inferences))

	for j, beta := range node.betaNodes {
		if beta.alphaNot && len(beta.parentNode.facts) == 0 {
			t.incoming[j] = &node.parentEngine.nullFact
		}
	}

	if f != nil {
		t.incoming[i] = f
	}

	return t, nil
}

func (node *pNode) removeToken(tok *token) (err error) {

	var i int
	var t *token
	var found bool

	for i, t = range node.tokens{

		if t == tok {
			found = true
			break
		}
	}
	if found {
		node.tokens[i] = node.tokens[len(node.tokens)-1]
		node.tokens[len(node.tokens)-1] = nil
		node.tokens = node.tokens[:len(node.tokens)-1]
	} else {
		return fmt.Errorf("removeToken: token not found")
	}
	return nil
}

func (node *pNode) activate(tok *token) (err error) {

	//first determine if token is complete
	for _, ptr := range tok.incoming {
		if ptr == nil {
			return nil
		}
	}

	//if so, fire off all inferences
	for i, inf := range node.inferences {

		f := Fact{}

		obj, ok := inf.ObjectId.(Variable)
		if ok { //then it's a variable
			tst := node.testNetwork[obj][0]
			if tst.objectElseValue {
				f.ObjectId = tok.incoming[tst.tokenIndex].ObjectId
			} else {
				f.ObjectId, ok = tok.incoming[tst.tokenIndex].Value.(string)
				if !ok {
					return fmt.Errorf("Inference failure: %s cannot be an ObjectId",reflect.TypeOf(tok.incoming[tst.tokenIndex].Value).Name())
				}
			}
		} else { //it should be a string value (but double-check)
			f.ObjectId, ok = inf.ObjectId.(string)
			if !ok {
				return fmt.Errorf("Inference failure: %s cannot be an ObjectId",reflect.TypeOf(inf.ObjectId).Name())
			}
		}

		f.Attribute = inf.Attribute

		val, ok := inf.Value.(Variable)
		if ok { //then it's a variable
			tst := node.testNetwork[val][0]
			if tst.objectElseValue {
				f.Value = tok.incoming[tst.tokenIndex].ObjectId
			} else {
				f.Value = tok.incoming[tst.tokenIndex].Value
			}
		} else {
			f.Value = inf.Value
		}

		node.parentEngine.pushAgenda(&f)
		tok.outgoing[i] = &f
	}
	return nil
}

func match(left interface{}, op Operator, right interface{}) (bool, error) {

	if left == nil && right == nil {
		if op == EQ {
			return true, nil
		} else if op == NE {
			return false, nil
		} else {
			return false, fmt.Errorf("match: cannot compare nil to nil using %s",op.String())
		}
	}

	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)

	if (leftValue.Kind() != rightValue.Kind()) {
		if op == EQ {
			return false, nil
		} else if op == NE {
			return true, nil
		} else {
			return false, fmt.Errorf("match: cannot compare %s to %s using %s",leftValue.Kind(),rightValue.Kind(),op.String())
		}
	}

	switch leftValue.Kind() {
	case reflect.String:
		switch op {
		case EQ:
			if leftValue.String() == rightValue.String() {
				return true, nil
			}
		case GE:
			if leftValue.String() >= rightValue.String() {
				return true, nil
			}
		case GT:
			if leftValue.String() > rightValue.String() {
				return true, nil
			}
		case LE:
			if leftValue.String() <= rightValue.String() {
				return true, nil
			}
		case LT:
			if leftValue.String() < rightValue.String() {
				return true, nil
			}
		case NE:
			if leftValue.String() != rightValue.String() {
				return true, nil
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch op {
		case EQ:
			if leftValue.Int() == rightValue.Int() {
				return true, nil
			}
		case GE:
			if leftValue.Int() >= rightValue.Int() {
				return true, nil
			}
		case GT:
			if leftValue.Int() > rightValue.Int() {
				return true, nil
			}
		case LE:
			if leftValue.Int() <= rightValue.Int() {
				return true, nil
			}
		case LT:
			if leftValue.Int() < rightValue.Int() {
				return true, nil
			}
		case NE:
			if leftValue.Int() != rightValue.Int() {
				return true, nil
			}
		}
	case reflect.Float32, reflect.Float64:
		switch op {
		case EQ:
			if leftValue.Float() == rightValue.Float() {
				return true, nil
			}
		case GE:
			if leftValue.Float() >= rightValue.Float() {
				return true, nil
			}
		case GT:
			if leftValue.Float() > rightValue.Float() {
				return true, nil
			}
		case LE:
			if leftValue.Float() <= rightValue.Float() {
				return true, nil
			}
		case LT:
			if leftValue.Float() < rightValue.Float() {
				return true, nil
			}
		case NE:
			if leftValue.Float() != rightValue.Float() {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("match: cannot compare %s",leftValue.Kind())
	}
	return false, nil
}

