// if using Jest:
// eslint-disable-next-line @typescript-eslint/consistent-type-imports
import { create as actualCreate, StateCreator } from 'zustand'
// if using Jest:
// import { StateCreator } from 'zustand';
// const { create: actualCreate } = jest.requireActual<typeof import('zustand')>('zustand');
import { act } from 'react'

// a variable to hold reset functions for all stores declared in the app
const storeResetFns = new Set<() => void>()

// when creating a store, we get its initial state, create a reset function and add it in the set
export const create =
  () =>
  <S>(createState: StateCreator<S>) => {
    const store = actualCreate(createState)
    const initialState = store.getState()
    // eslint-disable-next-line @typescript-eslint/no-confusing-void-expression
    storeResetFns.add(() => store.setState(initialState, true))
    return store
  }

// Reset all stores after each test run
beforeEach(() => {
  // eslint-disable-next-line @typescript-eslint/no-confusing-void-expression
  act(() => storeResetFns.forEach((resetFn) => { resetFn() }))
})
