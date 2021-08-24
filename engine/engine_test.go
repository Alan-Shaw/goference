package engine

import "math/rand"
import "testing"
import "time"

func TestDefine(t *testing.T) {

	var obj1 Variable = "variable1"

	simpleRule := Rule{
		Id: "simple-rule",
		LHS: []Condition{
			Condition{
				ObjectId:   obj1,
				Attribute:  "attribute1",
				Comparator: EQ,
				Value:      "value1",
			},
			Condition{
				ObjectId:   "object2",
				Attribute:  "attribute2",
				Comparator: GT,
				Value:      0.0,
			},
			Condition{
				ObjectId:   obj1,
				Attribute:  "attribute3",
				Comparator: LT,
				Value:      10,
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  obj1,
				Attribute: "attribute4",
				Value:     3.14,
			},
		},
	}

	testEngine := Engine{}

	err := testEngine.Define(simpleRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",simpleRule.Id,err)
	}

	if testEngine.alphaNetwork == nil {
		t.Errorf("No alpha network.")
		t.FailNow()
	}
	if len(testEngine.alphaNetwork) != 3 {
		t.Errorf("Alpha network has unexpected length: %d.", len(testEngine.alphaNetwork))
	}

	alphaCount := 0
	betaCount := 0
	for _, slc := range testEngine.alphaNetwork {

		for _, a := range slc {
			alphaCount++
			betaCount += len(a.betaNodes)
		}
	}
	if alphaCount != 3 {
		t.Errorf("Expected 3 alpha nodes, found %d", alphaCount)
	}
	if betaCount != 3 {
		t.Errorf("Expected 3 beta nodes, found %d", betaCount)
	}
}

func TestAlphaNodeMatch(t *testing.T) {

	testFact := Fact{
		ObjectId:  "0",
		Attribute: "test",
		Value:     nil,
	}

	testNode := alphaNode{
		attributeName: "test",
		objConstraint: "",
		compareTo:     nil,
	}

	test := "Float-Int-EQ"
	testFact.Value = 3.14
	testNode.comparator = EQ
	testNode.compareTo = 3
	matched, err := match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-Float-NE"
	testFact.Value = 3
	testNode.comparator = NE
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "String-EQ-Match"
	testFact.Value = "value1"
	testNode.comparator = EQ
	testNode.compareTo = "value1"
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "String-EQ-Mismatch"
	testFact.Value = "value2"
	testNode.comparator = EQ
	testNode.compareTo = "value3"
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "String-NE-Match"
	testFact.Value = "value4"
	testNode.comparator = NE
	testNode.compareTo = "value5"
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "String-NE-Mismatch"
	testFact.Value = "value6"
	testNode.comparator = NE
	testNode.compareTo = "value6"
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-EQ-Match"
	testFact.Value = 3
	testNode.comparator = EQ
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-EQ-Mismatch"
	testFact.Value = 3
	testNode.comparator = EQ
	testNode.compareTo = 4
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GE-Match1"
	testFact.Value = 3
	testNode.comparator = GE
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GE-Match2"
	testFact.Value = 4
	testNode.comparator = GE
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GE-Misatch"
	testFact.Value = 3
	testNode.comparator = GE
	testNode.compareTo = 4
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GT-Match"
	testFact.Value = 4
	testNode.comparator = GT
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GT-Mismatch1"
	testFact.Value = 3
	testNode.comparator = GT
	testNode.compareTo = 4
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-GT-Mismatch2"
	testFact.Value = 3
	testNode.comparator = GT
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LE-Match1"
	testFact.Value = 3
	testNode.comparator = LE
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LE-Match2"
	testFact.Value = 3
	testNode.comparator = LE
	testNode.compareTo = 4
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LE-Misatch"
	testFact.Value = 4
	testNode.comparator = LE
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LT-Match"
	testFact.Value = 3
	testNode.comparator = LT
	testNode.compareTo = 4
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LT-Mismatch1"
	testFact.Value = 4
	testNode.comparator = LT
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Int-LT-Mismatch2"
	testFact.Value = 3
	testNode.comparator = LT
	testNode.compareTo = 3
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-EQ-Match"
	testFact.Value = 3.14
	testNode.comparator = EQ
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-EQ-Mismatch"
	testFact.Value = 3.14
	testNode.comparator = EQ
	testNode.compareTo = 4.0
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GE-Match1"
	testFact.Value = 3.14
	testNode.comparator = GE
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GE-Match2"
	testFact.Value = 4.0
	testNode.comparator = GE
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GE-Misatch"
	testFact.Value = 3.14
	testNode.comparator = GE
	testNode.compareTo = 4.0
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GT-Match"
	testFact.Value = 4.0
	testNode.comparator = GT
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GT-Mismatch1"
	testFact.Value = 3.14
	testNode.comparator = GT
	testNode.compareTo = 4.0
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-GT-Mismatch2"
	testFact.Value = 3.14
	testNode.comparator = GT
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LE-Match1"
	testFact.Value = 3.14
	testNode.comparator = LE
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LE-Match2"
	testFact.Value = 3.14
	testNode.comparator = LE
	testNode.compareTo = 4.0
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LE-Misatch"
	testFact.Value = 4.0
	testNode.comparator = LE
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LT-Match"
	testFact.Value = 3.14
	testNode.comparator = LT
	testNode.compareTo = 4.0
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || !matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LT-Mismatch1"
	testFact.Value = 4.0
	testNode.comparator = LT
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}

	test = "Float-LT-Mismatch2"
	testFact.Value = 3.14
	testNode.comparator = LT
	testNode.compareTo = 3.14
	matched, err = match(testFact.Value, testNode.comparator, testNode.compareTo)
	if err != nil || matched {
		t.Errorf("%s test failed.", test)
	}
}

func TestEngine(t *testing.T) {

	var err error
	var nextRule Rule
	var testVar1 Variable
	var testVar2 Variable
	var testVar3 Variable

	testEngine := Engine{}

	testVar1 = "var1"
	testVar2 = "var2"

	nextRule = Rule{
		Id: "L1R1",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "testAttr1",
				Comparator: GT,
				Value:      17.1187,
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "testAttr1",
				Comparator: LT,
				Value:      17.11378,
			},
			Condition{
				ObjectId:   testVar1,
				Attribute:  "testAttr2",
				Comparator: GE,
				Value:      55.87,
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "testAttr2",
				Comparator: LE,
				Value:      61.922927,
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar1,
				Attribute: "inferAttr1",
				Value:     5.908,
			},
			Inference{
				ObjectId:  testVar2,
				Attribute: "inferAttr2",
				Value:     "symbolic",
			},
			Inference{
				ObjectId:  "constant1",
				Attribute: "required",
				Value:     "true",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	testVar1 = "var4"
	testVar2 = "var5"
	testVar3 = "var6"

	nextRule = Rule{
		Id: "L1R2",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "testAttr3",
				Comparator: EQ,
				Value:      "this one",
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "testAttr3",
				Comparator: EQ,
				Value:      testVar1,
			},
			Condition{
				ObjectId:   testVar3,
				Attribute:  "testAttr3",
				Comparator: EQ,
				Value:      testVar2,
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar1,
				Attribute: "inferAttr3",
				Value:     "symbolic",
			},
			Inference{
				ObjectId:  testVar3,
				Attribute: "inferAttr4",
				Value:     "almost ready",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	testVar1 = "var7"
	testVar2 = "var8"
	testVar3 = "intVal"

	nextRule = Rule{
		Id: "L1R3",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "testAttr6",
				Comparator: EQ,
				Value:      testVar3,
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "testAttr7",
				Comparator: EQ,
				Value:      testVar3,
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar1,
				Attribute: "inferAttr5",
				Value:     "always",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	testVar1 = "var3"
	testVar2 = "var4"

	nextRule = Rule{
		Id: "L2R1",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "inferAttr1",
				Comparator: EQ,
				Value:      5.908,
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "inferAttr2",
				Comparator: EQ,
				Value:      "symbolic",
			},
			Condition{
				ObjectId:   "constant1",
				Attribute:  "required",
				Comparator: EQ,
				Value:      "true",
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar1,
				Attribute: "penultimate",
				Value:     "ready",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	testVar1 = "var1"
	testVar2 = "var2"
	testVar3 = "var3"

	nextRule = Rule{
		Id: "L2R2",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "inferAttr3",
				Comparator: EQ,
				Value:      "symbolic",
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "inferAttr4",
				Comparator: EQ,
				Value:      "almost ready",
			},
			Condition{
				ObjectId:   testVar3,
				Attribute:  "inferAttr5",
				Comparator: EQ,
				Value:      "always",
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar3,
				Attribute: "penultimate",
				Value:     "set",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	testVar1 = "var4"
	testVar2 = "var5"

	nextRule = Rule{
		Id: "L3R1",
		LHS: []Condition{
			Condition{
				ObjectId:   testVar1,
				Attribute:  "penultimate",
				Comparator: EQ,
				Value:      "ready",
			},
			Condition{
				ObjectId:   testVar2,
				Attribute:  "penultimate",
				Comparator: EQ,
				Value:      "set",
			},
			Condition{
				NotExists:	true,
				ObjectId:   "",
				Attribute:  "is-deal-killer",
				Comparator: EQ,
				Value:      "true",
			},
		},
		RHS: []Inference{
			Inference{
				ObjectId:  testVar1,
				Attribute: "passed",
				Value:     "true",
			},
		},
	}

	err = testEngine.Define(nextRule)
	if err != nil {
		t.Errorf("Error defining rule %s: %s\n",nextRule.Id,err)
	}

	//it is important that the results be the same regardless of the order in which facts are asserted
	ind := [9]int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(9, func(i, j int) { ind[i], ind[j] = ind[j], ind[i] })

	//fact set 1:
	factSet := [9]Fact{ Fact{ObjectId: "set1obj1", Attribute: "testAttr1", Value: 18.123},
						Fact{ObjectId: "set1obj2", Attribute: "testAttr1", Value: 10.456},
						Fact{ObjectId: "set1obj1", Attribute: "testAttr2", Value: 55.87},
						Fact{ObjectId: "set1obj2", Attribute: "testAttr2", Value: 61.922927},
						Fact{ObjectId: "set1obj3", Attribute: "testAttr3", Value: "this one"},
						Fact{ObjectId: "set1obj4", Attribute: "testAttr3", Value: "set1obj3"},
						Fact{ObjectId: "set1obj5", Attribute: "testAttr3", Value: "set1obj4"},
						Fact{ObjectId: "set1obj6", Attribute: "testAttr6", Value: 42},
						Fact{ObjectId: "set1obj7", Attribute: "testAttr7", Value: 42},
				}
	for _, i := range ind {
		err = testEngine.Assert(factSet[i])
		if err != nil{
			t.Errorf(err.Error())
		}
	}

	var test string
	var expected int

	/***********************/
	test = "1 Fact Set 1"
	expected = 1
	result, err := testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}

	/***********************/
	test = "2 Fact Set 1"
	expected = 0
	err = testEngine.Retract(factSet[ind[0]])
	if err != nil{
		t.Errorf(err.Error())
	}
	result, err = testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}

	/***********************/
	test = "3 Fact Set 1"
	expected = 1
	err = testEngine.Assert(factSet[ind[0]])
	if err != nil{
		t.Errorf(err.Error())
	}
	result, err = testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}

	/***********************/
	test = "4 Fact Set 1"
	expected = 0
	err = testEngine.Assert(Fact{ObjectId: "any1", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	err = testEngine.Assert(Fact{ObjectId: "any2", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	err = testEngine.Assert(Fact{ObjectId: "any3", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	result, err = testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}

	/***********************/
	test = "5 Fact Set 1"
	expected = 1
	err = testEngine.Retract(Fact{ObjectId: "any2", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	err = testEngine.Retract(Fact{ObjectId: "any3", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	err = testEngine.Retract(Fact{ObjectId: "any1", Attribute: "is-deal-killer", Value: "true"})
	if err != nil{
		t.Errorf(err.Error())
	}
	result, err = testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(9, func(i, j int) { ind[i], ind[j] = ind[j], ind[i] })

	//fact set 2:
	factSet = [9]Fact{ Fact{ObjectId: "set2obj1", Attribute: "testAttr1", Value: 20.31},
						Fact{ObjectId: "set2obj2", Attribute: "testAttr1", Value: 9.582},
						Fact{ObjectId: "set2obj1", Attribute: "testAttr2", Value: 62.109},
						Fact{ObjectId: "set2obj2", Attribute: "testAttr2", Value: 50.824},
						Fact{ObjectId: "set2obj3", Attribute: "testAttr3", Value: "this one"},
						Fact{ObjectId: "set2obj4", Attribute: "testAttr3", Value: "set2obj3"},
						Fact{ObjectId: "set2obj5", Attribute: "testAttr3", Value: "set2obj4"},
						Fact{ObjectId: "set2obj6", Attribute: "testAttr6", Value: 42},
						Fact{ObjectId: "set2obj7", Attribute: "testAttr7", Value: 42},
				}
	for _, i := range ind {
		err = testEngine.Assert(factSet[i])
		if err != nil{
			t.Errorf(err.Error())
		}
	}

	test = "6 Fact Set 2"
	expected = 3
	result, err = testEngine.GetInferences("","passed")
	if err != nil{
		t.Errorf(err.Error())
	}
	if len(result) != expected {
		t.Errorf("Test %s: expected %d, got %d\n",test,expected,len(result))
	}
}
