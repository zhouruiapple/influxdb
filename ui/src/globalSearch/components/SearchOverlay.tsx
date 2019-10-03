// Libraries
import React, {FunctionComponent, ChangeEvent, useState} from 'react'
import {WithRouterProps} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
import {
  Form,
  Input,
  Overlay,
  ComponentSize,
  ButtonType,
  Radio,
  ComponentColor,
  SquareButton,
  IconFont,
  FlexBox,
  FlexDirection,
  JustifyContent,
  TextBlock,
} from '@influxdata/clockface'
import ResourcesList from './ResourcesList'

// Actions
import {getGlobalSearch} from 'src/globalSearch/actions'

interface OwnProps {}

interface DispatchProps {
  getGlobalSearch: typeof getGlobalSearch
}

type Props = OwnProps & WithRouterProps & DispatchProps

const SearchOverlay: FunctionComponent<Props> = props => {
  const [searchTerm, updateSearch] = useState('')
  const [activeType, updateActiveType] = useState('')

  const resourceTypes = [
    'bucket',
    'dashboards',
    'labels',
    'telegraf',
    'org',
    'user',
  ]

  const handleSearch = () => {
    props.getGlobalSearch(activeType)
  }

  const closeModal = () => {
    props.router.goBack()
  }

  const handleSearchInput = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value

    updateSearch(value)
  }

  return (
    <Overlay visible={true}>
      <Overlay.Container maxWidth={1000}>
        <Overlay.Header title="Search" onDismiss={closeModal} />
        <Overlay.Body>
          <Form onSubmit={handleSearch}>
            <FlexBox
              direction={FlexDirection.Row}
              stretchToFitWidth={true}
              justifyContent={JustifyContent.SpaceBetween}
              margin={ComponentSize.Large}
            >
              <Form.Element label="" errorMessage="">
                <Input
                  placeholder="Whatchu Need?"
                  name="search"
                  autoFocus={true}
                  value={searchTerm}
                  onChange={handleSearchInput}
                  testID="search-input"
                  size={ComponentSize.Large}
                />
              </Form.Element>
              <FlexBox.Child grow={0}>
                <Form.Element label="" errorMessage="">
                  <SquareButton
                    titleText="Search"
                    type={ButtonType.Submit}
                    color={ComponentColor.Default}
                    size={ComponentSize.Large}
                    icon={IconFont.Search}
                    testID="create-org-submit-button"
                  />
                </Form.Element>
              </FlexBox.Child>
            </FlexBox>
            <FlexBox
              direction={FlexDirection.Row}
              margin={ComponentSize.Large}
              justifyContent={JustifyContent.FlexStart}
              stretchToFitWidth={false}
            >
              <TextBlock text="Filter By Resource Type:" />
              <Radio>
                {resourceTypes.map(btn => (
                  <Radio.Button
                    key={btn}
                    id={btn}
                    active={btn === activeType}
                    value={btn}
                    titleText={btn}
                    onClick={value => updateActiveType(value)}
                  >
                    {btn}
                  </Radio.Button>
                ))}
              </Radio>
            </FlexBox>
          </Form>
        </Overlay.Body>
        <Overlay.Footer>
          <ResourcesList searchTerm={searchTerm} />
        </Overlay.Footer>
      </Overlay.Container>
    </Overlay>
  )
}

const mdtp = {
  getGlobalSearch: getGlobalSearch,
}

export default connect<{}, DispatchProps, {}>(
  null,
  mdtp
)(SearchOverlay)
