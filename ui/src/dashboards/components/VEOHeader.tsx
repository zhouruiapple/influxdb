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
  Input,
  Page,
  Button,
  Overlay,
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

const algorithms = [{ name: 'Facebook Prophet', id: 'sha256:f76d7d549549c4ea18b735e9373e2c4deb1392b93599e089430023dc3f088650' }]

class VEOHeader extends PureComponent<Props> {
  state = {
    forecastButtonEnabled: true,
    isOverlayVisible: false,
    destinationBucket: '',
    destinationMeasurement: '',
    serviceId: '',
    serviceName: '',
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
            <RagnarokAlgorithms algorithms={algorithms} onClick={this.displayOverlay} />
            <Overlay visible={this.state.isOverlayVisible}>
              <Overlay.Container maxWidth={600}>
                <Overlay.Header />
                <Overlay.Body>
                  <Input value={this.state.destinationBucket} placeholder="Destination Bucket" onChange={this.updateDestinationBucket} />
                  <Input value={this.state.destinationMeasurement} placeholder="Destination Measurement" onChange={this.updateDestinationMeasurement} />
                </Overlay.Body>
                <Overlay.Footer>
                  <Button text="Apply" onClick={this.applyAlgorithm} />
                  <Button text="Cancel" onClick={this.cancelAlgorithm} />
                </Overlay.Footer>
              </Overlay.Container>
            </Overlay>
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

  private updateDestinationBucket = (event) => {
    this.setState({ destinationBucket: event.target.value })
  }

  private updateDestinationMeasurement = (event) => {
    this.setState({ destinationMeasurement: event.target.value })
  }

  private cancelAlgorithm = () => {
     this.setState({isOverlayVisible: false})
  }

  private displayOverlay = ({id, name}) => {
    this.setState({isOverlayVisible: true, serviceId: id, serviceName: name})
  }

  private applyAlgorithm = async() => {
    console.log(this.state)
    this.setState({isOverlayVisible: false})
    this.forecast();
  }

  private forecast = async() => {
    this.setState({forecastButtonEnabled: false})

    // grab this on mount
    // const services = await listServices()
    // const {id: serviceId, name: serviceName} = services[0]

    const instance = await getInstance({name: this.state.serviceName, serviceId: this.state.serviceId})
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
