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
//import {RagnarokAlgorithms} from 'src/dashboards/utils/RagnarokAlgorithms'
import {RagnarokServicesDropdown} from 'src/dashboards/utils/RagnarokServicesDropdown'

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
    isProcessingOverlayVisible: false,
    destinationBucket: '',
    destinationMeasurement: '',
    serviceId: '',
    serviceName: '',
    services: null,
    params: [{name:"days",placeholder:"ph",value:"v"},{name:"age",placeholder:"ph",value:"v"}],
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
            {/*<RagnarokAlgorithms algorithms={algorithms} onClick={this.displayOverlay} />*/}
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
            <Overlay visible={this.state.isOverlayVisible}>
              <Overlay.Container maxWidth={600}>
                <Overlay.Header title={this.state.serviceName}/>
                <Overlay.Body>
                {this.state.params.map(r => (
                 [<Form.Label required={true} label="Form Label"/>,
                  <Input  value={r.value} placeholder={r.placeholder} onChange={(e)=>{this.updateParam(r.name,e)}} />,
                  <Form.HelpText text="help text"></Form.HelpText>]
                  ))}
              
                <Form.Divider lineColor={ComponentColor.Primary}/>
                <Form.Label required={true} label="Destination Bucket"/>
                <Input value={this.state.destinationBucket} placeholder="Destination Bucket" onChange={this.updateDestinationBucket} />
                <Form.HelpText text="The bucket to write resuls to"/>

                <Form.Divider lineColor={ComponentColor.Primary}/>
                <Form.Label required={true} label="Destination Measurement"/>
                <Input value={this.state.destinationMeasurement} placeholder="Destination Measurement" onChange={this.updateDestinationMeasurement} />
                <Form.HelpText text="The measurement to write resuls to"/>


                <Form.Divider lineColor={ComponentColor.Primary}/>
                <Form.Label required={false} label="Output Tags"/>
                <Input value="foo,bar" placeholder="Output Tags"/>
                <Form.HelpText text="Any additional tags to attach to the results"/>

                <Form.Divider lineColor={ComponentColor.Primary}/>
                <Form.Label required={false} label="Repeat Every"/>
                <Input value="0s" placeholder="How often to repeat"/>
                <Form.HelpText text="If required, how often should this action to be repeated?"/>

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

  private updateParam = (name, event) => {
    console.log(name,"changed to",event.target.value)
    let p = this.state.params.slice(0) // otherwise nasty cycles happen
    p.forEach(element => {
      if (element.name == name) {
        element.value = event.target.value
      }
    });
    this.setState({params:p})
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
