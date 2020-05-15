// Libraries
import React, {PureComponent, MouseEvent} from 'react'
import {connect} from 'react-redux'

// Components
import RenamablePageTitle from 'src/pageLayout/components/RenamablePageTitle'
import {
  SquareButton,
  ComponentColor,
  ComponentSize,
  ComponentStatus,
  IconFont,
  Page,
  Button,
  Overlay
} from '@influxdata/clockface'
import VisOptionsButton from 'src/timeMachine/components/VisOptionsButton'
import ViewTypeDropdown from 'src/timeMachine/components/view_options/ViewTypeDropdown'

// ragnarok stuff
import {addQuery, editActiveQueryAsFlux, setActiveQueryText} from 'src/timeMachine/actions'
import {saveAndExecuteQueries} from 'src/timeMachine/actions/queries'
import {getActiveQuery} from 'src/timeMachine/selectors'
import {RagnarokAlgorithms} from 'src/dashboards/utils/RagnarokAlgorithms'

// Constants
import {
  DEFAULT_CELL_NAME,
  CELL_NAME_MAX_LENGTH,
} from 'src/dashboards/constants/index'

import {getInstance, listServices, runWhenComplete, startForecasting} from 'src/dashboards/utils/ragnarok'

interface Props {
  name: string
  onSetName: (name: string) => void
  onCancel: () => void
  onSave: () => void
}

const saveButtonClass = 'veo-header--save-cell-button'

const algorithms = [{ name: 'Facebook Prophet', id: 'asdf' }]

class VEOHeader extends PureComponent<Props> {
  state = {
    forecastButtonEnabled: true
  }

  get forecastButtonStatus() {
    if (this.state.forecastButtonEnabled) {
      return ComponentStatus.Default
    }
    return ComponentStatus.Loading
  }

  public render() {
    const {name, onSetName, onCancel, onSave} = this.props
    return (
      <>
        <Page.Header fullWidth={true}>
          <RenamablePageTitle
            name={name}
            onRename={onSetName}
            placeholder={DEFAULT_CELL_NAME}
            maxLength={CELL_NAME_MAX_LENGTH}
            onClickOutside={this.handleClickOutsideTitle}
          />
        </Page.Header>
        <Page.ControlBar fullWidth={true}>
          <Page.ControlBarLeft>
            <ViewTypeDropdown />
            <VisOptionsButton />
            <RagnarokAlgorithms algorithms={algorithms} onClick={this.handleAlgorithmSelect} />
          </Page.ControlBarLeft>
          <Page.ControlBarRight>
            <SquareButton
              icon={IconFont.Remove}
              onClick={onCancel}
              size={ComponentSize.Small}
              testID="cancel-cell-edit--button"
            />
            <SquareButton
              className={saveButtonClass}
              icon={IconFont.Checkmark}
              color={ComponentColor.Success}
              size={ComponentSize.Small}
              onClick={onSave}
              testID="save-cell--button"
            />
          </Page.ControlBarRight>
        </Page.ControlBar>
      </>
    )
  }


  private handleAlgorithmSelect = async({id, name}) => {
    console.log('name, id', name, id)
    this.forecast();
  }

  private forecast = async() => {
    this.setState({forecastButtonEnabled: false})
    const services = await listServices()
    const {id: serviceId, name: serviceName} = services[0]

    const instance = await getInstance({name: serviceName, serviceId})
    const {id: instanceId} = instance

    const activityId = await startForecasting(instanceId, this.props.activeQuery.text)

    runWhenComplete(activityId, () => {
      this.setState({forecastButtonEnabled: true})
      const forecastQuery =
`from(bucket: "ds-bucket")
   |> range(start: -15)
   |> filter(fn: (r) => r["_measurement"] == "forecast")
   |> filter(fn: (r) => r["_field"] == "yhat_lower" or r["_field"] == "yhat_upper")`

      this.props.addQuery()
      this.props.editActiveQueryAsFlux()
      this.props.setActiveQueryText(forecastQuery)
      this.props.saveAndExecuteQueries()
    })
  }

  private handleClickOutsideTitle = (e: MouseEvent<HTMLElement>) => {
    const {onSave} = this.props
    const target = e.target as HTMLButtonElement

    if (!target.className.includes(saveButtonClass)) {
      return
    }

    onSave()
  }
}

const mstp = (state: AppState): StateProps => {
  const activeQuery = getActiveQuery(state)

  return {activeQuery}
}


const mdtp = {
  addQuery,
  editActiveQueryAsFlux,
  saveAndExecuteQueries,
  setActiveQueryText,
}

export default connect(
  mstp,
  mdtp
)(VEOHeader)
