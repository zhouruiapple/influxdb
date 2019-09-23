import {ComponentSize} from '@influxdata/clockface'

export type ActionTypes =
  | ToggleShowVariablesControlsAction
  | TogglePageSizeAction

interface ToggleShowVariablesControlsAction {
  type: 'TOGGLE_SHOW_VARIABLES_CONTROLS'
}

export interface TogglePageSizeAction {
  type: 'TOGGLE_PAGE_SIZE'
  payload: {pageSize: ComponentSize}
}

export const toggleShowVariablesControls = (): ToggleShowVariablesControlsAction => ({
  type: 'TOGGLE_SHOW_VARIABLES_CONTROLS',
})

export const togglePageSize = (pageSize): TogglePageSizeAction => ({
  type: 'TOGGLE_PAGE_SIZE',
  payload: {pageSize},
})
