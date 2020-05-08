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
  Button
} from '@influxdata/clockface'
import VisOptionsButton from 'src/timeMachine/components/VisOptionsButton'
import ViewTypeDropdown from 'src/timeMachine/components/view_options/ViewTypeDropdown'

// Constants
import {
  DEFAULT_CELL_NAME,
  CELL_NAME_MAX_LENGTH,
} from 'src/dashboards/constants/index'

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
    let servicesResponse
    try {
      servicesResponse = await fetch('http://localhost:8081/api/services')
    } catch (err) {
      console.log('error', err)
    }
    if (!servicesResponse.ok) {
      alert('error')
      return
    }

    const [meta] = await servicesResponse.json()
    const {id: serviceId, name: serviceName} = meta


    console.log('meta', meta)

    // curl -XPOST http://localhost:8081/api/instances --data '{"name":"MyFB1","serviceId":"sha256:b824898cc847a69b8d514cb241e427dd4e75f295dfb3c5cc86913998068d795c"}' --header "Content-Type: application/json"

    const instances = {
      name: serviceName,
      serviceId
    }

    let instancesResponse
    try {
       instancesResponse = await fetch('http://localhost:8081/api/instances', {
        method: 'POST',
        body: JSON.stringify(instances),
        headers: {
          'Content-Type': 'application/json',
        },
      })
    } catch (error) {
      console.error(error)
    }

    if (!instancesResponse.ok) {
      alert('error')
      return
    }

    const instance = await instancesResponse.json()
    const {id: instanceId} = instance

    const body = {
      instanceId,
      operationName: 'Forecast',
      inputQuery: 'from(bucket: "ds-bucket") |> range(start: -15y) |> filter(fn: (r) => r["_measurement"] == "historical") |> filter(fn: (r) => r["_field"] == "value")',
      outputDatabase: 'ds-bucket',
      outputMeasurement: 'forecasting',
      params: {Days: '365'},
    }

    let activitiesResponse
    try {
       activitiesResponse = await fetch('http://localhost:8081/api/activities', {
        method: 'POST',
        body: JSON.stringify(body),
        headers: {
          'Content-Type': 'application/json',
        },
      })
    } catch (error) {
      console.error(error)
    }

    if (!activitiesResponse.ok) {
      alert('error')
      return
    }

    const {activityId} = await activitiesResponse.json()
    console.log('result!', activityId)
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
