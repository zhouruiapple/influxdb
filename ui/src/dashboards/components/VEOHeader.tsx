// Libraries
import React, {PureComponent, MouseEvent} from 'react'

// Components
import RenamablePageTitle from 'src/pageLayout/components/RenamablePageTitle'
import {
  SquareButton,
  ComponentColor,
  ComponentSize,
  IconFont,
  Page,
  Button,
  Overlay
} from '@influxdata/clockface'
import VisOptionsButton from 'src/timeMachine/components/VisOptionsButton'
import ViewTypeDropdown from 'src/timeMachine/components/view_options/ViewTypeDropdown'

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

class VEOHeader extends PureComponent<Props> {
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
            <Button
              color={ComponentColor.Primary}
              onClick={this.forecast}
              size={ComponentSize.Small}
              text="Forecast"
            />
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

  private forecast = async() => {
    const services = await listServices()
    const {id: serviceId, name: serviceName} = services[0]


    console.log('services', services)

    // curl -XPOST /ragnarok/instances --data '{"name":"MyFB1","serviceId":"sha256:b824898cc847a69b8d514cb241e427dd4e75f295dfb3c5cc86913998068d795c"}' --header "Content-Type: application/json"

    const instance = await getInstance({name: serviceName, serviceId})
    const {id: instanceId} = instance
    console.log('instance', instance)

    const body = {
      instanceId,
      operationName: 'Forecast',
      inputQuery: 'from(bucket: "ds-bucket") |> range(start: -15y) |> filter(fn: (r) => r["_measurement"] == "historical") |> filter(fn: (r) => r["_field"] == "value")',
      outputDatabase: 'ds-bucket',
      outputMeasurement: 'forecasting',
      params: {Days: '365'},
    }

    const inputQuery = 'from(bucket: "ds-bucket") |> range(start: -15y) |> filter(fn: (r) => r["_measurement"] == "historical") |> filter(fn: (r) => r["_field"] == "value")'

    const activityId = await startForecasting(instanceId, inputQuery)
    console.log('result!', activityId)

    runWhenComplete(activityId, () => {
      console.log('callback, hi!')
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

export default VEOHeader
