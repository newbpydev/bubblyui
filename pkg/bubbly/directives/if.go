package directives

// IfDirective implements conditional rendering with ElseIf and Else support.
//
// The If directive provides a declarative way to conditionally render content
// based on boolean conditions. It supports chaining multiple conditions via
// ElseIf and provides a fallback via Else.
//
// # Basic Usage
//
//	If(condition, func() string {
//	    return "Condition is true"
//	}).Render()
//
// # With Else
//
//	If(condition, func() string {
//	    return "True branch"
//	}).Else(func() string {
//	    return "False branch"
//	}).Render()
//
// # With ElseIf Chain
//
//	If(status == "loading",
//	    func() string { return "Loading..." },
//	).ElseIf(status == "error",
//	    func() string { return "Error occurred" },
//	).ElseIf(status == "empty",
//	    func() string { return "No data" },
//	).Else(func() string {
//	    return "Data loaded"
//	}).Render()
//
// # Nested If
//
//	If(outerCondition, func() string {
//	    return If(innerCondition, func() string {
//	        return "Both true"
//	    }).Else(func() string {
//	        return "Outer true, inner false"
//	    }).Render()
//	}).Render()
//
// # Type Safety
//
// All branch functions must return strings. The directive is type-safe and
// will catch type mismatches at compile time.
//
// # Performance
//
// The directive evaluates conditions lazily - only the branch that matches
// is executed. This makes it efficient even with expensive render functions.
//
// # Purity
//
// The directive is pure - it has no side effects and always produces the same
// output for the same input. Branch functions should also be pure for predictable
// behavior.
type IfDirective struct {
	condition      bool
	thenBranch     func() string
	elseIfBranches []ElseIfBranch
	elseBranch     func() string
}

// ElseIfBranch represents a single ElseIf condition and its associated branch.
//
// This type is used internally by IfDirective to store chained ElseIf conditions.
// Each ElseIfBranch contains a boolean condition and a function to execute if
// that condition is true (and all previous conditions were false).
type ElseIfBranch struct {
	condition bool
	branch    func() string
}

// If creates a new conditional directive with the given condition and then branch.
//
// The If function is the entry point for conditional rendering. It evaluates the
// condition and, if true, executes the then function. If false, it checks any
// ElseIf branches or the Else branch.
//
// Parameters:
//   - condition: Boolean expression to evaluate
//   - then: Function to execute if condition is true
//
// Returns:
//   - *IfDirective: A new If directive that can be chained with ElseIf/Else
//
// Example:
//
//	If(user.IsAdmin(), func() string {
//	    return "Admin Panel"
//	}).Else(func() string {
//	    return "User Panel"
//	}).Render()
//
// The returned directive implements ConditionalDirective, allowing method chaining
// for ElseIf and Else branches.
func If(condition bool, then func() string) *IfDirective {
	return &IfDirective{
		condition:      condition,
		thenBranch:     then,
		elseIfBranches: []ElseIfBranch{},
		elseBranch:     nil,
	}
}

// ElseIf adds an additional conditional branch to the directive chain.
//
// This method allows chaining multiple conditions, where each condition is
// evaluated in order until one is true. If this ElseIf's condition is true
// and all previous conditions were false, the provided then function is executed.
//
// Parameters:
//   - condition: Boolean expression to evaluate
//   - then: Function to execute if condition is true and all previous conditions were false
//
// Returns:
//   - ConditionalDirective: Self reference for method chaining
//
// Example:
//
//	If(score >= 90, func() string { return "A" }).
//	    ElseIf(score >= 80, func() string { return "B" }).
//	    ElseIf(score >= 70, func() string { return "C" }).
//	    Else(func() string { return "F" }).
//	    Render()
//
// ElseIf branches are evaluated in the order they are added. The first matching
// condition's branch is executed, and subsequent branches are skipped.
func (d *IfDirective) ElseIf(condition bool, then func() string) ConditionalDirective {
	d.elseIfBranches = append(d.elseIfBranches, ElseIfBranch{
		condition: condition,
		branch:    then,
	})
	return d
}

// Else provides a fallback branch when all previous conditions are false.
//
// This method completes the conditional chain by providing a default branch
// that executes when neither the initial If condition nor any ElseIf conditions
// are true. Only one Else can be specified per conditional chain.
//
// Parameters:
//   - then: Function to execute if all previous conditions were false
//
// Returns:
//   - ConditionalDirective: Self reference for method chaining (allows Render())
//
// Example:
//
//	If(hasData, func() string {
//	    return renderData()
//	}).Else(func() string {
//	    return "No data available"
//	}).Render()
//
// If Else is not called and all conditions are false, Render() returns an empty string.
func (d *IfDirective) Else(then func() string) ConditionalDirective {
	d.elseBranch = then
	return d
}

// Render executes the directive logic and returns the resulting string output.
//
// This method evaluates the conditional chain in order:
//  1. If the main condition is true, execute the then branch
//  2. Otherwise, check each ElseIf condition in order
//  3. If an ElseIf condition is true, execute its branch
//  4. If all conditions are false, execute the Else branch (if present)
//  5. If no Else branch and all conditions false, return empty string
//
// Returns:
//   - string: The rendered output from the first matching branch, or empty string
//
// Example:
//
//	result := If(status == "loading",
//	    func() string { return "Loading..." },
//	).ElseIf(status == "error",
//	    func() string { return "Error!" },
//	).Else(func() string {
//	    return "Ready"
//	}).Render()
//
// The method is pure and idempotent - calling it multiple times with the same
// state produces the same result. Only the matching branch function is executed,
// making it efficient even with expensive render functions.
func (d *IfDirective) Render() string {
	// Check main condition
	if d.condition {
		return d.thenBranch()
	}

	// Check ElseIf branches in order
	for _, branch := range d.elseIfBranches {
		if branch.condition {
			return branch.branch()
		}
	}

	// Execute Else branch if present
	if d.elseBranch != nil {
		return d.elseBranch()
	}

	// No conditions met and no Else branch
	return ""
}
