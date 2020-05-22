// Libraries
import React, {FC, createRef} from 'react'

// Component
import SnippitTooltipContents from 'src/timeMachine/components/snippitsToolbar/SnippitTooltipContents'
import {
  Popover,
  PopoverPosition,
  PopoverInteraction,
  Appearance,
  Button,
  ComponentSize,
  ComponentColor,
  ConfirmationButton,
  IconFont,
  FlexBox,
} from '@influxdata/clockface'

// Types
import {Snippit} from 'src/types'
import {connect} from 'react-redux'

// Actions
import {deleteSnippit} from 'src/snippits/actions/thunks'


interface DispatchProps {
  onDelete: typeof deleteSnippit
}

interface Props {
  snippit: Snippit
  onClickSnippit: (snippit: string) => void
  testID?: string
}

const defaultProps = {
  testID: 'flux-snippit',
}

const SnippitItem: FC<Props & DispatchProps> = ({snippit, onClickSnippit, onDelete, testID}) => {
  const snippitRef = createRef<HTMLDListElement>()
  const handleClickSnippit = () => {
    onClickSnippit(snippit.code)
  }

  return (
    <>
      <Popover
        appearance={Appearance.Outline}
        color={ComponentColor.Secondary}
        enableDefaultStyles={false}
        position={PopoverPosition.ToTheLeft}
        triggerRef={snippitRef}
        showEvent={PopoverInteraction.Hover}
        hideEvent={PopoverInteraction.Hover}
        distanceFromTrigger={8}
        testID="toolbar-popover"
        contents={() => <SnippitTooltipContents snippit={snippit} />}
      />
      <dd
        ref={snippitRef}
        data-testid={`flux--${testID}`}
        className="flux-toolbar--list-item flux-toolbar--snippit"
      >
        <code>{snippit.name}</code>
        <FlexBox className="flux-toolbar--injector" margin={ComponentSize.Small}>
          <Button
            testID={`flux--${testID}--inject`}
            text="Inject"
            onClick={handleClickSnippit}
            size={ComponentSize.ExtraSmall}
            color={ComponentColor.Secondary}
          />
          <ConfirmationButton
            size={ComponentSize.ExtraSmall}
            icon={IconFont.Trash}
            color={ComponentColor.Danger}
            confirmationLabel="Delete your life's work?"
            confirmationButtonText="OMG Yes!"
            confirmationButtonColor={ComponentColor.Danger}
            onConfirm={() => onDelete(snippit.id)} />
        </FlexBox>
      </dd>
    </>
  )
}

SnippitItem.defaultProps = defaultProps

const mdtp = {
  onDelete: deleteSnippit
}

export default connect<{}, DispatchProps, {}>(
  null,
  mdtp
)(SnippitItem)
