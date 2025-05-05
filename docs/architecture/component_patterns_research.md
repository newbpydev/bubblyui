# React and Solid.js Component Patterns Research

## Overview

This document examines component patterns from React and Solid.js for adaptation to BubblyUI's Go-based terminal UI framework. The goal is to identify patterns that can be effectively translated to Go while maintaining the component-based, reactive architecture these frameworks are known for.

## React Component Patterns

### 1. Functional Components

React's functional components are pure functions that accept props and return JSX. They form the foundation of modern React development.

**Example:**
```jsx
function Button({ label, onClick }) {
  return (
    <button onClick={onClick}>
      {label}
    </button>
  );
}
```

**Key Concepts:**
- Pure rendering functions with no internal state
- Receive data via props
- Return UI description (JSX)
- Can use hooks for state and effects

### 2. Hooks for State Management

React hooks allow functional components to use state and other React features.

**Example:**
```jsx
function Counter() {
  const [count, setCount] = useState(0);
  
  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </div>
  );
}
```

**Key Concepts:**
- `useState` for local component state
- `useEffect` for side effects and lifecycle events
- `useContext` for consuming context
- `useReducer` for complex state logic
- `useMemo` and `useCallback` for optimization

### 3. Context for Global State

React Context provides a way to share values between components without explicitly passing props.

**Example:**
```jsx
const ThemeContext = React.createContext('light');

function App() {
  return (
    <ThemeContext.Provider value="dark">
      <ThemedButton />
    </ThemeContext.Provider>
  );
}

function ThemedButton() {
  const theme = useContext(ThemeContext);
  return <button className={theme}>Themed Button</button>;
}
```

## Solid.js Component Patterns

### 1. Fine-grained Reactivity

Solid.js uses a fine-grained reactivity system based on signals, which only update the DOM elements that depend on changed values.

**Example:**
```jsx
function Counter() {
  const [count, setCount] = createSignal(0);
  
  return (
    <div>
      <p>Count: {count()}</p>
      <button onClick={() => setCount(count() + 1)}>Increment</button>
    </div>
  );
}
```

**Key Concepts:**
- Signals are getter/setter pairs
- Only DOM nodes dependent on signals update (not the whole component)
- No virtual DOM diffing
- Reactive computations via `createMemo` and `createEffect`

### 2. Reactive Primitives

Solid.js provides several reactive primitives:

**Example:**
```jsx
function User(props) {
  // Derived state with createMemo
  const fullName = createMemo(() => `${props.firstName} ${props.lastName}`);
  
  // Side effects with createEffect
  createEffect(() => {
    console.log(`Name changed to ${fullName()}`);
  });
  
  return <div>Hello, {fullName()}</div>;
}
```

**Key Concepts:**
- `createSignal` for reactive state
- `createMemo` for derived/computed values
- `createEffect` for side effects
- `createResource` for async data fetching

### 3. Props and Stores

Solid.js has a unique approach to props and stores:

**Example:**
```jsx
// Reactive store
const [state, setState] = createStore({
  user: { name: "John", age: 25 },
  settings: { theme: "dark" }
});

// Component using nested reactivity
function UserProfile() {
  return (
    <div>
      <h1>{state.user.name}</h1>
      <p>Age: {state.user.age}</p>
      <button onClick={() => setState("user", "age", a => a + 1)}>
        Happy Birthday
      </button>
    </div>
  );
}
```

## Adaptation Considerations for Go

### 1. Component Interface

In Go, we can define a `Component` interface that captures the essential behaviors of React/Solid components:

```go
type Component interface {
    Initialize() error
    Update(msg tea.Msg) (tea.Cmd, error)
    Render() string
    Dispose() error
}
```

### 2. Reactive State (Signals)

We can implement signal-like reactivity in Go:

```go
type Signal[T any] struct {
    value      T
    subscribers map[string]func(T)
    mutex      sync.RWMutex
}

func NewSignal[T any](initialValue T) *Signal[T] {
    return &Signal[T]{
        value:      initialValue,
        subscribers: make(map[string]func(T)),
    }
}

// Getter with dependency tracking
func (s *Signal[T]) Value() T {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    // Track dependency here
    
    return s.value
}

// Setter with notification
func (s *Signal[T]) SetValue(newValue T) {
    s.mutex.Lock()
    
    // Skip if value hasn't changed
    if reflect.DeepEqual(s.value, newValue) {
        s.mutex.Unlock()
        return
    }
    
    s.value = newValue
    subscribers := s.subscribers
    s.mutex.Unlock()
    
    // Notify subscribers
    for _, notify := range subscribers {
        notify(newValue)
    }
}
```

### 3. Props System

Props in Go could be implemented as struct fields:

```go
type ButtonProps struct {
    Label    string
    OnClick  func()
    Disabled bool
}

type Button struct {
    props ButtonProps
    // Internal state
}

func NewButton(props ButtonProps) *Button {
    return &Button{
        props: props,
    }
}
```

### 4. Component Composition

Component composition in Go would differ from React/Solid but can achieve similar goals:

```go
type Container struct {
    props       ContainerProps
    children    []Component
    innerWidth  *Signal[int]
    innerHeight *Signal[int]
}

func (c *Container) Render() string {
    var result strings.Builder
    
    // Render container borders
    
    // Render children
    for _, child := range c.children {
        childOutput := child.Render()
        // Position and append child output
        result.WriteString(childOutput)
    }
    
    return result.String()
}
```

## Next Steps

Based on this research, we should:

1. Define the concrete `Component` interface for BubblyUI
2. Implement signal-based reactivity
3. Design a consistent parent-child communication pattern
4. Create a proof-of-concept with simple components
