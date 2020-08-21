// Libraries
import React, {FC, PureComponent} from 'react'
import {Switch, Route} from 'react-router-dom'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'
import DemoDataIndex from 'src/writeData/components/demoData/DemoDataIndex'
import WriteDataDetailsView from 'src/writeData/components/WriteDataDetailsView'
import WriteDataHelper from 'src/writeData/components/WriteDataHelper'

// Constants
import {ORGS, ORG_ID, DEMO_DATA} from 'src/shared/constants/routes'
import {WRITE_DATA_DEMO_DATA_SECTION} from 'src/writeData/constants/contentDemoData'

const demoDataPath = `/${ORGS}/${ORG_ID}/load-data/${DEMO_DATA}`

const DemoDataDetailsPage: FC = () => {
  return (
    <WriteDataDetailsView section={WRITE_DATA_DEMO_DATA_SECTION}>
      <WriteDataHelper />
    </WriteDataDetailsView>
  )
}

@ErrorHandling
class DemoDataPage extends PureComponent {
  public render() {
    const {children} = this.props

    return (
      <>
        <Switch>
          <Route path={demoDataPath} exact component={DemoDataIndex} />
          <Route
            path={`${demoDataPath}/:contentID`}
            component={DemoDataDetailsPage}
          />
        </Switch>
        {children}
      </>
    )
  }
}

export default DemoDataPage
