// Libraries
import React, {FC} from 'react'

// Components
import {Page} from '@influxdata/clockface'
import {ResultsProvider} from 'src/notebooks/context/results'
import {RefProvider} from 'src/notebooks/context/refs'
import CurrentNotebook from 'src/notebooks/context/notebook.current'
import {ScrollProvider} from 'src/notebooks/context/scroll'
import Header from 'src/notebooks/components/header'
import PipeList from 'src/notebooks/components/PipeList'
import MiniMap from 'src/notebooks/components/minimap/MiniMap'
import BucketProvider from 'src/notebooks/context/buckets'

// NOTE: uncommon, but using this to scope the project
// within the page and not bleed it's dependancies outside
// of the feature flag
import 'src/notebooks/style.scss'

const NotebookPage: FC = () => {
  return (
    <CurrentNotebook>
      <ResultsProvider>
        <RefProvider>
          <ScrollProvider>
            <BucketProvider>
              <Page titleTag="Flows">
                <Header />
                <Page.Contents
                  fullWidth={true}
                  scrollable={false}
                  className="notebook-page"
                >
                  <div className="notebook">
                    <MiniMap />
                    <PipeList />
                  </div>
                </Page.Contents>
              </Page>
            </BucketProvider>
          </ScrollProvider>
        </RefProvider>
      </ResultsProvider>
    </CurrentNotebook>
  )
}

export default NotebookPage
