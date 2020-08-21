// Libraries
import React, {FC} from 'react'

// Components
import WriteDataIndexView from 'src/writeData/components/WriteDataIndexView'

// Constants
import {WRITE_DATA_DEMO_DATA_SECTION} from 'src/writeData/constants/contentDemoData'

const DemoDataIndex: FC = ({children}) => {
  return (
    <>
      <WriteDataIndexView content={WRITE_DATA_DEMO_DATA_SECTION} />
      {children}
    </>
  )
}

export default DemoDataIndex
