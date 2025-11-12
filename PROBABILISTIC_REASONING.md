# Probabilistic Multi-Response Reasoning Protocol

## Purpose
This protocol enhances AI decision-making quality by generating multiple solution candidates, evaluating their probabilities, and selecting the optimal approach through systematic analysis.

## When to Apply
Use this protocol for:
- **Complex architectural decisions** (component design, API design, system architecture)
- **High-stakes implementations** (core framework features, breaking changes)
- **Ambiguous requirements** (multiple valid interpretations)
- **Performance-critical code** (optimization strategies)
- **Uncertain situations** (when confidence is low on the best approach)

**Do NOT use for:**
- Simple, well-defined tasks (file reads, basic edits)
- Trivial decisions with obvious solutions
- Time-sensitive quick fixes

## Protocol Steps

### Step 1: Generate 5 Alternative Solutions
For the given problem, generate exactly 5 distinct solution approaches. Each solution should:
- Be meaningfully different from the others
- Be technically viable
- Address the core problem
- Have clear trade-offs

### Step 2: Evaluate Each Solution
For each solution, provide:

1. **Confidence Score** (0-100%): Likelihood this solution is optimal
2. **Evaluation Criteria Scores** (each 0-10):
   - **Correctness**: Does it fully solve the problem?
   - **Performance**: Efficiency and resource usage
   - **Maintainability**: Code clarity, future extensibility
   - **Type Safety**: Strong typing, compile-time guarantees
   - **Testability**: Ease of writing comprehensive tests
   - **Idiomatic**: Follows Go/project conventions
   - **Complexity**: Lower is better (cognitive load)

3. **Composite Score**: Weighted average of criteria
   ```
   Score = (Correctness × 0.25) + (Performance × 0.15) +
           (Maintainability × 0.20) + (Type Safety × 0.15) +
           (Testability × 0.15) + (Idiomatic × 0.05) +
           (Complexity × 0.05)
   ```

4. **Pros**: Key advantages (2-3 points)
5. **Cons**: Key disadvantages (2-3 points)
6. **Risks**: Potential failure modes

### Step 3: Present Solutions

```markdown
## Solution Analysis

### Solution 1: [Brief Name]
**Confidence**: 85%
**Composite Score**: 8.2/10

**Approach**: [1-2 sentence description]

**Scores**:
- Correctness: 9/10
- Performance: 8/10
- Maintainability: 8/10
- Type Safety: 9/10
- Testability: 8/10
- Idiomatic: 8/10
- Complexity: 7/10

**Pros**:
- Strong type safety with generics
- Clear separation of concerns
- Well-tested pattern in the codebase

**Cons**:
- Slightly more boilerplate
- Requires additional test setup

**Risks**:
- May introduce breaking changes to existing API

---

[Repeat for Solutions 2-5]
```

### Step 4: Selection Rationale
After presenting all 5 solutions:

1. **Rank solutions** by composite score
2. **Identify top 2-3** candidates
3. **Compare trade-offs** between top candidates
4. **Make final selection** with clear justification
5. **Address why alternatives were rejected**

### Step 5: Proceed with Implementation
Once the optimal solution is selected, proceed with implementation using the standard workflow (TDD, specs, etc.).

## Example Format

```markdown
# Problem: Implement state persistence for component lifecycle

## Solution 1: WeakMap-based State Cache
**Confidence**: 75% | **Score**: 7.8/10
[Details...]

## Solution 2: Interface-based State Manager
**Confidence**: 90% | **Score**: 8.9/10
[Details...]

## Solution 3: Closure-based Private State
**Confidence**: 60% | **Score**: 6.5/10
[Details...]

## Solution 4: Struct Embedding Pattern
**Confidence**: 70% | **Score**: 7.2/10
[Details...]

## Solution 5: Context-based State Injection
**Confidence**: 85% | **Score**: 8.5/10
[Details...]

---

## Selection: Solution 2 (Interface-based State Manager)

**Rationale**:
- Highest composite score (8.9/10) and confidence (90%)
- Best balance of type safety, maintainability, and testability
- Aligns with existing component architecture patterns
- Clear interface contracts reduce cognitive complexity

**Why not others**:
- Solution 1: WeakMap not idiomatic in Go, memory concerns
- Solution 3: Closures harder to test, less explicit
- Solution 4: Embedding creates tight coupling
- Solution 5: Context overhead for this use case

**Proceeding with**: Interface-based State Manager approach
```

## Confidence Calibration Guidelines

- **90-100%**: High certainty, well-established pattern, minimal risk
- **70-89%**: Good confidence, minor trade-offs or unknowns
- **50-69%**: Moderate confidence, significant trade-offs
- **30-49%**: Low confidence, substantial risks or uncertainties
- **0-29%**: Very low confidence, likely suboptimal

## Integration with Project Workflow

1. **When uncertain**: Apply this protocol BEFORE writing code
2. **Document decision**: Add analysis to relevant spec or design doc
3. **Review selection**: Use code review skill to validate choice
4. **Iterate if needed**: If solution fails, revisit alternative options

## Activation

To activate this protocol, include in your prompt:
```
Apply the Probabilistic Multi-Response Reasoning Protocol from PROBABILISTIC_REASONING.md
```

Or simply:
```
Generate 5 solutions with probabilities and select the best approach
```

---

**Remember**: This protocol adds thinking overhead. Use it strategically for decisions that warrant deep analysis, not for every minor choice.
