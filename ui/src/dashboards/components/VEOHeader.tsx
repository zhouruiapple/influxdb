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
  TextBlock,
  Form,

} from '@influxdata/clockface'
import VisOptionsButton from 'src/timeMachine/components/VisOptionsButton'
import ViewTypeDropdown from 'src/timeMachine/components/view_options/ViewTypeDropdown'

// ragnarok stuff
import {addQuery, editActiveQueryAsFlux, setActiveQueryText} from 'src/timeMachine/actions'
import {saveAndExecuteQueries} from 'src/timeMachine/actions/queries'
import {getActiveQuery} from 'src/timeMachine/selectors'
import {getTimeRange} from 'src/dashboards/selectors'
import {RagnarokServicesDropdown} from 'src/dashboards/utils/RagnarokServicesDropdown'
import {RagnarokActionParametersForm} from 'src/dashboards/utils/RagnarokActionParametersForm'

//import { uniqueNamesGenerator, Config, adjectives, colors, animals } from 'unique-names-generator';


// Constants
import {
  DEFAULT_CELL_NAME,
  CELL_NAME_MAX_LENGTH,
} from 'src/dashboards/constants/index'

import {getInstance, listServices, runWhenComplete, executeAction, listInstances} from 'src/dashboards/utils/ragnarok'
import { ActionTypes } from '../actions/ranges'
import QueryTab from 'src/timeMachine/components/QueryTab'

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
    instances: null,
    service: null,
    instance: null,
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
    listInstances().then((instances)=>{
      this.setState({instances:instances})
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
            <RagnarokServicesDropdown services={this.state.services} instances={this.state.instances} onClick={this.displayOverlay} />
            <Overlay visible={this.state.isProcessingOverlayVisible}>
              <Notification
                horizontalAlignment={Alignment.Center}
                  key="k"
                  id="i"
                  gradient = {Gradients.DefaultDark}
                  size={ComponentSize.Large}
                >
                  {this.state.action && <TextBlock text={`Executing ${this.state.action.name}`}/>}
                  {this.state.instance && <TextBlock text={`Instance: ${this.state.instance.name}`}/>}
                  {this.state.service && <TextBlock text={`Service: ${this.state.service.name}`}/>}

                  <SpinnerContainer
                    loading={RemoteDataState.Loading}
                    spinnerComponent={<TechnoSpinner />}
                  >
            </SpinnerContainer>
                </Notification>
            </Overlay>

          {this.state.action != null && <RagnarokActionParametersForm isVisible={this.state.isOverlayVisible} service={this.state.service} action={this.state.action} instance={this.state.instance} onCancel={this.cancelAlgorithm} onApply={this.applyAlgorithm}/>}
            
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
     this.setState({serviceId: null, serviceName: null, service:null, action:null, outputTags:null,repeat:null, instance:null})
     this.setState({isOverlayVisible: false})
  }

  private displayOverlay = ({id, name, service, action, instance}) => {
    console.log("service selected",service)
    console.log("action selected",action)
    console.log("instance selected",instance)
    let ot = null
    if (action.output != null) {
      ot = action.output.defaultTags
    }
    this.setState({serviceId: id, serviceName: name, service:service, action:action, outputTags:ot,repeat:'0s', instance:instance})
    this.setState({isOverlayVisible: true})
  }

  private applyAlgorithm = async(obj: any) => {
    console.log(this.state,obj)
    this.setState({isOverlayVisible: false, isProcessingOverlayVisible: true})
    this.executeServiceAction(obj);
  }

  private executeServiceAction = async(obj: any) => {


    //       query.text.includes('v.timeRangeStart') ||
    //  query.text.includes('v.timeRangeStop')
    console.log("active time range",this.props.timeRange)

    // fix the query up with
    // timeRange.lower
    // timeRange.upper  - mauy be null!
    // range(start: v.timeRangeStart, stop: v.timeRangeStop)

    let query = this.props.activeQuery.text
    if (query.includes('range(start: v.timeRangeStart, stop: v.timeRangeStop)')) {
      if (this.props.timeRange.lower != null && this.props.timeRange.upper != null) {
        query = query.replace('range(start: v.timeRangeStart, stop: v.timeRangeStop)',
        `range(start: ${this.props.timeRange.lower}, stop: ${this.props.timeRange.upper})`)
      } else if (this.props.timeRange.duration != null) {
        query = query.replace('range(start: v.timeRangeStart, stop: v.timeRangeStop)',
        `range(start: -${this.props.timeRange.duration})`)
      }
    }


    const instanceName: string = obj.instanceName

    console.log('instance name',instanceName)


   // const instanceName = `auto-${Math.random()}`
    let instance = this.state.instance
    if (instance == null) {
      instance = await getInstance({name: instanceName, serviceId: this.state.serviceId})
    }
    const {id: instanceId} = instance

    const activityId = await executeAction(instanceId, this.state.action.name, query, obj)

    runWhenComplete(activityId, (activity: any) => {
      this.setState({forecastButtonEnabled: true, isProcessingOverlayVisible: false})
        if (this.state.action.output!=null) {


          const forecastQuery = this.generateActivityResultQuery(activity)

          //const forecastQuery =
    //`from(bucket: "forecasting-bucket")
    //   |> range(start: -15)
    //   |> filter(fn: (r) => r["_measurement"] == "forecast")
    //   |> filter(fn: (r) => r["_field"] == "yhat_lower" or r["_field"] == "yhat_upper")`

          this.props.addQuery()
          this.props.editActiveQueryAsFlux()
          this.props.setActiveQueryText(forecastQuery)
          this.props.saveAndExecuteQueries()
        } else {
          console.log("action completed but no action.output so no query generated")
        }
        this.setState({serviceId: null, serviceName: null, service:null, action:null, outputTags:null,repeat:null, instance:null})
      }
    )
  }

  private generateActivityResultQuery (activity: any) {
    let query = ''
    if (activity.resultInfo != null) {
      const min = new Date(activity.resultInfo.minTimestamp)
      const max = new Date(activity.resultInfo.maxTimestamp)
      query = `from(bucket:"${activity.resultInfo.bucket}")\n`
      query = query + `|> range(start:${min.toISOString()}, stop:${max.toISOString()})\n`
      query = query + `|> filter(fn: (r) => r["_measurement"] == "${activity.resultInfo.measurement}")\n`
      if (this.state.action != null && this.state.action.output && this.state.action.output.presentationHints) {
        let first = true
        for (const df of this.state.action.output.presentationHints.defaultFields) {
          if (!first) {
            query = query + ` or r["_field"] == "${df}"`
          }
          if (first) {
            query = query + `|> filter(fn: (r) => r["_field"] == "${df}"`
            first = false
          } 
        }
        if (!first) {
          query = query + ")\n"
        }
        for (const ot of activity.resultInfo.outputTags) {
          query = query + `|> filter(fn: (r) => r["${ot.name}"] == "${ot.value}")\n`  
        }
      }
      return query
    }

    console.log("query",query)
    // from bucket
    // range lowest to highest time range from activity
    // filter measurement name
    // filter result fields  
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
  const timeRange = getTimeRange(state)

  return {activeQuery,timeRange}
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
