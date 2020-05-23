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
  Notification,
  Gradients,
  Alignment,
  SpinnerContainer,
  TechnoSpinner,
  RemoteDataState,
  Form,

} from '@influxdata/clockface'
import VisOptionsButton from 'src/timeMachine/components/VisOptionsButton'
import ViewTypeDropdown from 'src/timeMachine/components/view_options/ViewTypeDropdown'


// ragnarok stuff
import {addQuery, editActiveQueryAsFlux, setActiveQueryText} from 'src/timeMachine/actions'
import {saveAndExecuteQueries} from 'src/timeMachine/actions/queries'
import {getActiveQuery} from 'src/timeMachine/selectors'
import {RagnarokServicesDropdown} from 'src/dashboards/utils/RagnarokServicesDropdown'
import {RagnarokActionParametersForm} from 'src/dashboards/utils/RagnarokActionParametersForm'


// Constants
import {
  DEFAULT_CELL_NAME,
  CELL_NAME_MAX_LENGTH,
} from 'src/dashboards/constants/index'

import {getInstance, listServices, runWhenComplete, startForecasting} from 'src/dashboards/utils/ragnarok'
import { ActionTypes } from '../actions/ranges'

interface Props {
  name: string
  onSetName: (name: string) => void
  onCancel: () => void
  onSave: () => void
}

const saveButtonClass = 'veo-header--save-cell-button'

class VEOHeader extends PureComponent<Props> {
  state = {
    forecastButtonEnabled: true,
    isOverlayVisible: false,
    isProcessingOverlayVisible: false,
    serviceId: '',
    serviceName: '',
    services: null,
    service: null,
    action: null,
  }

  timerID: any

  componentDidMount() {
    this.timerID = setInterval(
      () => this.tick(),
      2000
    );
  }

  componentWillUnmount() {
    clearInterval(this.timerID);
  }

  tick() {
    listServices().then((services)=>{
      this.setState({services:services})
    })
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
            <RagnarokServicesDropdown services={this.state.services} onClick={this.displayOverlay} />
            <Overlay visible={this.state.isProcessingOverlayVisible}>
              <Notification
                horizontalAlignment={Alignment.Center}
                  key="k"
                  id="i"
                  gradient = {Gradients.DefaultDark}
                  size={ComponentSize.Large}
                >
                  <span className="notification--message">Processing {this.state.serviceName}</span>
                  <SpinnerContainer
                    loading={RemoteDataState.Loading}
                    spinnerComponent={<TechnoSpinner />}
                  >
            </SpinnerContainer>
                </Notification>
            </Overlay>

          {this.state.action != null && <RagnarokActionParametersForm isVisible={this.state.isOverlayVisible} service={this.state.service} action={this.state.action} onCancel={this.cancelAlgorithm} onApply={this.applyAlgorithm}/>}
            
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

  private cancelAlgorithm = () => {
     this.setState({isOverlayVisible: false})
  }

  private displayOverlay = ({id, name, service, action}) => {
    console.log("setting outputTags",action.output.defaultTags)
    this.setState({isOverlayVisible: true, serviceId: id, serviceName: name, service:service, action:action, outputTags:action.output.defaultTags,repeat:'0s'})
  }

  private applyAlgorithm = async(obj: any) => {
    console.log(this.state,obj)
    this.setState({isOverlayVisible: false, isProcessingOverlayVisible: true})
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
      this.setState({forecastButtonEnabled: true, isProcessingOverlayVisible: false})
      const forecastQuery =
`from(bucket: "forecasting-bucket")
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
