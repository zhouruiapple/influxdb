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
} from '@influxdata/clockface'

// Types
import {Snippit} from 'src/types'

interface Props {
  snippit: Snippit
  onClickSnippit: (snippit: string) => void
  testID?: string
}

const defaultProps = {
  testID: 'flux-snippit',
}

const ToolbarSnippit: FC<Props> = ({snippit, onClickSnippit, testID}) => {
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
        <Button
          testID={`flux--${testID}--inject`}
          text="Inject"
          onClick={handleClickSnippit}
          size={ComponentSize.ExtraSmall}
          className="flux-toolbar--injector"
          color={ComponentColor.Secondary}
        />
      </dd>
    </>
  )
}

ToolbarSnippit.defaultProps = defaultProps

export default ToolbarSnippit
