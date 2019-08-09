// Libraries
import React, {useState, FC} from 'react'
import {Input} from '@influxdata/clockface'
import {isEqual} from 'lodash'

// Actions
import {
  search,
  clearSearch,
} from 'src/eventViewer/components/EventViewer.reducer'

// Utils
import {useDebouncedValue} from 'src/shared/utils/useDebouncedValue'
import {useMountedEffect} from 'src/shared/utils/useMountedEffect'

// Types
import {EventViewerChildProps, SearchExpr} from 'src/eventViewer/types'

const SEARCH_DELAY_MS = 500

const searchExprForTerm = (term: string): SearchExpr | null => {
  if (term.trim() === '') {
    return null
  }

  return term
}

type Props = EventViewerChildProps & {
  placeholder?: string
}

const LogsSearchBar: FC<Props> = ({state, dispatch, loadRows}) => {
  const [searchTerm, setSearchTerm] = useState<string>('')
  const debouncedSearchTerm = useDebouncedValue(searchTerm, SEARCH_DELAY_MS)
  const searchExpr = searchExprForTerm(debouncedSearchTerm)

  useMountedEffect(() => {
    if (searchExpr && !isEqual(searchExpr, state.searchExpr)) {
      search(state, dispatch, loadRows, searchExpr)
    } else if (!isEqual(searchExpr, state.searchExpr)) {
      clearSearch(state, dispatch, loadRows)
    }
  }, [state, dispatch, loadRows, searchExpr])

  return (
    <Input
      className="logs-search-bar"
      placeholder="Search..."
      value={searchTerm}
      onChange={e => setSearchTerm(e.target.value)}
    />
  )
}

export default LogsSearchBar
