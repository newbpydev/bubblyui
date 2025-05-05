# Communication Flow Diagrams

This document illustrates the communication patterns in BubblyUI through detailed flow diagrams, showing how components interact with each other and how data and events propagate through the component tree.

## 1. Unidirectional Data Flow

The core principle of BubblyUI's communication model is unidirectional data flow. Data flows down the component tree via props, while events flow up via callbacks.

```
┌───────────────────────────────────────────────────┐
│                                                   │
│                   Application                     │
│                                                   │
│    ┌─────────────────┐      ┌─────────────────┐   │
│    │                 │      │                 │   │
│    │     State       ├─────►│     Actions     │   │
│    │                 │      │                 │   │
│    └─────────────────┘      └─────────────────┘   │
│           │                          ▲            │
│           │                          │            │
│           ▼                          │            │
│    ┌─────────────────────────────────────────┐    │
│    │                                         │    │
│    │            Root Component               │    │
│    │                                         │    │
│    └─────────────────────────────────────────┘    │
│           │                          ▲            │
│           │                          │            │
└───────────┼──────────────────────────┼────────────┘
            │                          │
            │   Props Flow Down        │  Events Flow Up
            │                          │
            ▼                          │
┌────────────────────────────────────────────────────┐
│                                                    │
│                Parent Component                    │
│                                                    │
│    ┌─────────────────┐          ┌──────────────┐   │
│    │                 │          │              │   │
│    │  Props (static) │          │Event Handlers│   │
│    │                 │          │              │   │
│    └─────────────────┘          └──────────────┘   │
│           │                            ▲           │
│           │                            │           │
└───────────┼────────────────────────────┼───────────┘
            │                            │
            │                            │
            ▼                            │
┌────────────────────────────────────────────────────┐
│                                                    │
│                 Child Component                    │
│                                                    │
│    ┌─────────────────┐          ┌──────────────┐   │
│    │                 │          │              │   │
│    │     Render()    │          │    Events    │   │
│    │                 │          │              │   │
│    └─────────────────┘          └──────────────┘   │
│                                                    │
└────────────────────────────────────────────────────┘
```

## 2. Signal-Based Reactive Updates

The reactive nature of BubblyUI is powered by signals that automatically propagate changes through the component tree.

```
┌──────────────────────────────────────────────┐
│                                              │
│              Parent Component                │
│                                              │
│     ┌───────────────────┐                    │
│     │                   │                    │
│     │   count Signal    ├────────┐           │
│     │    value: 0       │        │           │
│     │                   │        │           │
│     └───────────────────┘        │           │
│               │                  │           │
└───────────────┼──────────────────┼───────────┘
                │                  │
                │                  │ Signal Passed
                │                  │ as Prop
                │                  │
                │                  ▼
                │      ┌───────────────────────────┐
                │      │                           │
                │      │     Child Component       │
                │      │                           │
                │      │   ┌───────────────────┐   │
                │      │   │                   │   │
Signal          │      │   │   count (prop)    │   │
Update          │      │   │    value: 0       │   │
                │      │   │                   │   │
                │      │   └───────────────────┘   │
                │      │                           │
                │      └───────────────────────────┘
                │                    │
                │                    │ Access Signal
                │                    │ During Render
                │                    │
                │                    ▼
                │      ┌───────────────────────────┐
                │      │                           │
                │      │  Dependency tracking      │
                │      │  registers render         │
                │      │  function as              │
                │      │  dependent of signal      │
                │      │                           │
                │      └───────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────┐
│                                              │
│           Signal Value Change                │
│           count.SetValue(1)                  │
│                                              │
└──────────────────────────────────────────────┘
                │
                │
                ▼
┌──────────────────────────────────────────────┐
│                                              │
│     Notification to all dependents           │
│     including Child Component                │
│                                              │
└──────────────────────────────────────────────┘
                │
                │
                ▼
┌──────────────────────────────────────────────┐
│                                              │
│     Child Component re-renders with          │
│     new value from signal: 1                 │
│                                              │
└──────────────────────────────────────────────┘
```

## 3. Event Callbacks Flow

Events generated from user interaction or internal state changes flow up the component tree through callback functions.

```
┌───────────────────────────────────────────────┐
│                                               │
│                 User Event                    │
│            (Click, Key Press, etc.)           │
│                                               │
└───────────────────────────────────────────────┘
                      │
                      ▼
┌───────────────────────────────────────────────┐
│                                               │
│              Child Component                  │
│                                               │
│  ┌─────────────────────────────────────┐      │
│  │                                     │      │
│  │  Event Handler (e.g., handleClick)  │      │
│  │                                     │      │
│  └─────────────────────────────────────┘      │
│                    │                          │
└────────────────────┼──────────────────────────┘
                     │
                     │ Call props.OnClick()
                     │
                     ▼
┌───────────────────────────────────────────────┐
│                                               │
│             Parent Component                  │
│                                               │
│  ┌─────────────────────────────────────┐      │
│  │                                     │      │
│  │  Event Handler (e.g., handleSubmit) │      │
│  │                                     │      │
│  └─────────────────────────────────────┘      │
│                    │                          │
└────────────────────┼──────────────────────────┘
                     │
                     │ Update state and/or call props.OnSubmit()
                     │
                     ▼
┌───────────────────────────────────────────────┐
│                                               │
│            Grandparent Component              │
│                                               │
│  ┌─────────────────────────────────────┐      │
│  │                                     │      │
│  │            State Update             │      │
│  │                                     │      │
│  └─────────────────────────────────────┘      │
│                    │                          │
└────────────────────┼──────────────────────────┘
                     │
                     │ Signal Updates
                     │
                     ▼
┌───────────────────────────────────────────────┐
│                                               │
│        Reactive UI Updates Propagate          │
│         From State Change Signals             │
│                                               │
└───────────────────────────────────────────────┘
```

## 4. Context-Based Communication

For scenarios where props drilling would be impractical, context provides a way to share data across component subtrees.

```
┌────────────────────────────────────────────────┐
│                                                │
│               Root Component                   │
│                                                │
│  ┌────────────────────────────────────────┐    │
│  │                                        │    │
│  │         Theme Context Provider         │    │
│  │                                        │    │
│  └────────────────────────────────────────┘    │
│                      │                         │
└──────────────────────┼─────────────────────────┘
                       │
                       │ Context Available
                       │ to All Descendants
                       │
                       ▼
     ┌─────────────────────────────────────┐
     │                                     │
     │          Parent Component           │
     │     (doesn't use theme context)     │
     │                                     │
     └─────────────────────────────────────┘
                       │
                       │ Normal Props Flow
                       │
                       ▼
┌────────────────────────────────────────────────┐
│                                                │
│             Child Component A                  │
│                                                │
│  ┌────────────────────────────────────────┐    │
│  │                                        │    │
│  │     Get Theme Context and Use It       │    │
│  │                                        │    │
│  └────────────────────────────────────────┘    │
│                                                │
└────────────────────────────────────────────────┘
                       │
                       │ Normal Props Flow
                       │
                       ▼
┌────────────────────────────────────────────────┐
│                                                │
│            Child Component B                   │
│                                                │
│  ┌────────────────────────────────────────┐    │
│  │                                        │    │
│  │    Also Gets Theme Context Directly    │    │
│  │                                        │    │
│  └────────────────────────────────────────┘    │
│                                                │
└────────────────────────────────────────────────┘

```

## 5. Component Lifecycle Communication

Components communicate during initialization, updates, and disposal through lifecycle events.

```
┌───────────────────────────────────────────────────────┐
│                                                       │
│                      Mount Phase                      │
│                                                       │
├─────────────────────────┬─────────────────────────────┤
│                         │                             │
│  ┌─────────────────┐    │    ┌─────────────────────┐  │
│  │                 │    │    │                     │  │
│  │  Initialize()   │────┼───>│  OnMount Effects    │  │
│  │                 │    │    │                     │  │
│  └─────────────────┘    │    └─────────────────────┘  │
│                         │                ┃            │
└─────────────────────────┼────────────────┃────────────┘
                          │                ┃
                          │                ┃ Set Up
                          │                ┃ Subscriptions
┌─────────────────────────┼────────────────▼─────────────┐
│                         │                              │
│                      Update Phase                      │
│                                                        │
├─────────────────────────┬──────────────────────────────┤
│                         │                              │
│  ┌─────────────────┐    │    ┌──────────────────────┐  │
│  │                 │    │    │                      │  │
│  │    Update()     │────┼───>│  OnUpdate Effects    │  │
│  │                 │    │    │                      │  │
│  └─────────────────┘    │    └──────────────────────┘  │
│          ┃              │                │             │
│          ┃              │                │             │
│          ▼              │                │             │
│  ┌─────────────────┐    │                │             │
│  │                 │    │                │             │
│  │    Render()     │<───┼────────────────┘             │
│  │                 │    │                              │
│  └─────────────────┘    │                              │
│                         │                              │
└─────────────────────────┼──────────────────────────────┘
                          │
                          │
┌─────────────────────────┼───────────────────────────────┐
│                         │                               │
│                     Unmount Phase                       │
│                                                         │
├─────────────────────────┬───────────────────────────────┤
│                         │                               │
│  ┌─────────────────┐    │    ┌─────────────────────┐    │
│  │                 │    │    │                     │    │
│  │    Dispose()    │────┼───>│  Cleanup Effects    │    │
│  │                 │    │    │                     │    │
│  └─────────────────┘    │    └─────────────────────┘    │
│                         │                │              │
└─────────────────────────┼────────────────┼──────────────┘
                          │                │
                          │                ▼
                          │     ┌────────────────────┐
                          │     │                    │
                          │     │ Remove Subscribers │
                          │     │                    │
                          │     └────────────────────┘
                          │
```

## 6. Component Tree Update Flow

When state changes occur, updates flow through the component tree in a specific order to ensure consistency.

```
                 ┌─────────────────┐
                 │                 │
                 │  Signal Change  │
                 │                 │
                 └─────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────┐
│                                                    │
│                Dependency Graph                    │
│                                                    │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐      │
│   │Signal A │────>│Computed │────>│Effect C │      │
│   └─────────┘     │Value B  │     └─────────┘      │
│       │           └─────────┘         ▲            │
│       │                 │             │            │
│       │                 ▼             │            │
│       │           ┌─────────┐         │            │
│       └──────────>│Effect D │─────────┘            │
│                   └─────────┘                      │
│                                                    │
└────────────────────────┬───────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────┐
│                                                    │
│               Update Scheduling                    │
│                                                    │
│   ┌─────────────────────────────────┐              │
│   │                                 │              │
│   │  Batch Updates (16ms window)    │              │
│   │                                 │              │
│   └─────────────────────────────────┘              │
│                    │                               │
└────────────────────┼───────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────┐
│                                                    │
│          Component Tree Updating                   │
│                                                    │
│  ┌─────────────┐                                   │
│  │             │                                   │
│  │    Root     │                                   │
│  │             │                                   │
│  └─────────────┘                                   │
│        │                                           │
│    ┌───┴───┐                                       │
│    │       │                                       │
│    ▼       ▼                                       │
│ ┌─────┐ ┌─────┐      Update Order:                 │
│ │  A  │ │  B  │      1. Computed Values            │
│ └─────┘ └─────┘      2. Effects                    │
│    │       │         3. Parent Components          │
│    │       │         4. Child Components           │
│    ▼       ▼                                       │
│ ┌─────┐ ┌─────┐                                    │
│ │  C  │ │  D  │                                    │
│ └─────┘ └─────┘                                    │
│                                                    │
└────────────────────────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────┐
│                                                    │
│               Final Rendering                      │
│                                                    │
│  ┌────────────────────────────────────┐            │
│  │                                    │            │
│  │  Compose all component outputs     │            │
│  │  into final terminal display       │            │
│  │                                    │            │
│  └────────────────────────────────────┘            │
│                                                    │
└────────────────────────────────────────────────────┘
```

## 7. Event Bubbling and Capturing

For certain types of events, BubblyUI supports event bubbling and capturing phases similar to the DOM event model.

```
┌────────────────────────────────────────────────────────┐
│                                                        │
│                   Event Capturing Phase                │
│              (Events travel down the tree)             │
│                                                        │
└────────────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────┐
│                                              │
│                Root Component                │
│                                              │
│  onCapture listener can intercept events     │
│  before they reach target                    │
│                                              │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                                              │
│              Parent Component                │
│                                              │
│  onCapture listener can intercept events     │
│  before they reach target                    │
│                                              │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                                              │
│          Target Component (Button)           │
│                                              │
│  Event occurs here (e.g., click)             │
│  Event handlers execute                      │
│                                              │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────┐
│                                                        │
│                   Event Bubbling Phase                 │
│              (Events travel up the tree)               │
│                                                        │
└────────────────────────────────────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                                              │
│              Parent Component                │
│                                              │
│  onBubble listener receives events           │
│  that have bubbled up from children          │
│                                              │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                                              │
│                Root Component                │
│                                              │
│  onBubble listener receives events           │
│  that have bubbled up from all descendants   │
│                                              │
└──────────────────────┬───────────────────────┘
                       │
                       │
                       ▼
                Event Processed

Event can be stopped at any point using:
- StopPropagation(): Prevents further bubbling
- PreventDefault(): Cancels default action
```

These diagrams illustrate the various communication patterns in the BubblyUI architecture. They provide a visual guide to how components interact, how data flows, and how the reactive system propagates updates through the application.
