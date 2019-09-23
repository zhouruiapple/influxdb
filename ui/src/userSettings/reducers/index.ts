import {ActionTypes} from 'src/userSettings/actions'
import {ComponentSize} from '@influxdata/clockface'

export interface UserSettingsState {
  showVariablesControls: boolean
  pageSize: ComponentSize
}

export const initialState = (): UserSettingsState => ({
  showVariablesControls: true,
  pageSize: ComponentSize.Large,
})

export const userSettingsReducer = (
  state = initialState(),
  action: ActionTypes
): UserSettingsState => {
  switch (action.type) {
    case 'TOGGLE_SHOW_VARIABLES_CONTROLS':
      return {...state, showVariablesControls: !state.showVariablesControls}
    case 'TOGGLE_PAGE_SIZE':
      const {pageSize} = action.payload
      return {...state, pageSize}
    default:
      return state
  }
}
