# goference
> An inference engine written in Go.

An inference engine (also known as a rules engine) allows one to define logical rules, add facts, and retrieve inferences that follow from those rules and facts. It is a means of adding logical reasoning to software.

In its current form, goference is not intended to be compiled into a standalone program or to provide a full programming environment like CLIPS or OPS5. However, it could be embedded within a larger Go application by a Go developer who has no qualms about defining rules using Go.

## Motivation

It has long been a goal of mine to try implementing the Rete algorithm, and Go seemed like an excellent tool for the job. However, when you set out to code something, you find out how much you really know about it, and I realized there were gaps in my understanding that I couldn't quite fill from the literature I had (mainly how to get identical results no matter in what order facts arrive). Instead of giving up, I just thought about it and filled in the functionality as best as I could. I do not claim that this is a particularly good or efficient implementation of Rete; it may not even be properly called Rete at all. I focused primarily on getting it to work while avoiding the temptation to prematurely optimize. And it does work, for my limited test cases, at least.

## Concepts

### Facts

The goference engine is based on one of the simplest fact representations: object-attribute-value (OAV). The object and attribute slots are always strings, but the value can also be an integer or floating point number. There is no requirement that the object id be unique. It is up to the knowledge engineer (i.e. rules programmer) to determine what goes there. This allows for multiple values, as in this example:

| Object     | Attribute   | Value         |
| :--------- | :---------- | :------------ |
| patientXYZ | has-symptom | fever         |
| patientXYZ | has-symptom | cough         |
| patientXYZ | has-symptom | headache      |
| patientXYZ | has-symptom | upset stomach |

Facts are asserted (i.e. added) one by one into the inference engine, which applies them to previously added rules.

### Rules

A rule is an IF-THEN statement. The IF side is called the left-hand side (LHS) or antecedent. The THEN side is called the right-hand side (RHS) or consequent.

### Conditions

The LHS of a rule is built out of one or more conditions. A condition defines a set of facts that matches it. It must specify a single, specific attribute. However, it may either specify a specific object id or accept any object id (meaning no restriction on object id). There is much more flexibility in the value comparison. Values may be strings, integers, or floating point numbers, and they may be compared with the full range of common operators: equality (EQ), inequality (NE), greater than (GT), greater than or equal to (GE), less than (LT), less than or equal to (LE).  

### Inferences

The RHS of a rule is built out of one or more inferences. In the context of goference, an inference is simply the assertion of a new fact. It is a fact asserted by the engine itself (back into itself) as opposed to being asserted externally.

### Quantification

The default condition has an implied existential quantifier (logic symbol ∃) and can be translated into English as "there exists one or more facts that meet this condition." 

However, conditions have a boolean flag that allows one to negate this and get the meaning: "no facts exists that meet this condition."

If you have had a logic course, you might be wondering, what about the universal quantifier (logic symbol ∀)? In the current version, there isn't one. However, you can work around this by negating it. For example, instead of a rule such as "if ALL participants are over 21, alcohol may be provided," you could write a rule such as "if ANY participant is under 21, alcohol will NOT be provided."

### Disjunction

Conditions are conjunctive. If A and B are both conditions in the LHS of a rule, this can be translated as "IF A AND B." The current version of goference does not support disjunction within a single rule, e.g. "IF A OR B." However, a rule such as "IF A OR B THEN C" could be rewritten as two rules: "IF A THEN C" and "IF B THEN C" to achieve the same result.

### Variables

A condition may contain a variable in the object id slot and/or the value slot (but not the attribute slot). These are not variables in the algebraic sense; they are not intended to express complex relationships. They are better thought of as tags. Their purpose is simply to bind facts from separate conditions together. For example, if these three conditions appear together in the LHS of the same rule:

|    | Object    | Value     |
| :- | :-------- | :-------- |
| 1  | variable1 |           |
| 2  | variable2 | variable1 |
| 3  |           | variable2 |

The rule will not fire unless object1 = value2, object2 = value3, object2 ≠ value2, and object1 ≠ value3.

Also note that when a variable appears in the value slot of a condition, the operator must be EQ. This is to emphasize that the condition does not constrain the value in this case, the binding *across* conditions does.

Variables are scoped to the rule in which they are found. There is no binding between separate rules.

Variables cannot be used with the negative quantifier (see above).

A variable *can* be extended into one or more inferences in the RHS of the same rule. This allows for dynamic inferences that assert new facts with values based on the specific facts that matched the rule. The same inference can then assert *different* facts (or more likely, the same fact about different objects). However, it is an error for a variable to appear for the first time in an inference, because there is no value to refer back to.

## Usage

The first thing you must do is create an empty engine:

`testEngine := Engine{}`

Next, you should declare any variables that you want to use in conditions.

A variable can be any string, because Variable is a custom type wrapped around string. This type has special meaning to the engine, though, so it cannot be used where an ordinary string is expected.

`var obj1 Variable = "variable1"`

A rule is a struct with three fields: an id string, a non-empty slice of conditions, and a non-empty slice of inferences. Defining a rule is simply a matter of first defining the rule structure with literals:

```

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
				NotExists:  true,
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
```

And then passing it to the Define() method of the engine:

`err = testEngine.Define(simpleRule)`

You should define all rules in the engine prior to adding the first fact. The engine will not stop you from defining further rules, but you may get incorrect results.

Adding facts is similar to adding rules but simpler. First define the fact with a literal, then pass it to the Assert() method:

```
	testFact := Fact{
		ObjectId:  "123A",
		Attribute: "test",
		Value:     "some value",
	}

	err = testEngine.Assert(testFact)
```
If you assert the same fact again, the engine will ignore it. The engine will also ignore "irrelevant" facts, i.e. facts that do not match any conditions. It does not save them in memory.

After each fact assertion, the internal state of the engine may change. You can check for inferences with the GetInferences() method. It takes two arguments, one for object id and one for attribute. It will return any inferences that have fired and that match. If either argument is the empty string, that one will be ignored. If both are empty, it will return all inferences that have fired.

`result, err = testEngine.GetInferences("456B","passed")`

Facts can be removed with the Retract() method, causing any and all inferences that depended on it to be rolled back. Facts do not have unique keys, so how does the engine know which fact you are retracting? It matches all three slots against the facts it has in memory and retracts the one that matches. If it doesn't find one, it takes no action.

`err = testEngine.Retract(testFact)`

After retracting one or more facts, a call to GetInferences() may reveal that some inferences that you retrieved earlier are no longer there.

## Bugs

These are inevitable, especially in something as complex as this. If you run into any, let me know.

Keep in mind, though, that poorly written rules can also cause bugs and that these are not the fault of the engine. Knowledge engineering is no trivial task, but it is perhaps preferable to extremely complex conventional code.

## Licensing

This project is licensed under the MIT License.

MIT © 2021 Alan D. Shaw

