import React, {FC, useEffect, useState, useCallback} from 'react'
import {connect, ConnectedProps} from 'react-redux'

// Actions
import {getBuckets} from 'src/buckets/actions/thunks'

// Selectors
import {getSortedBuckets} from 'src/buckets/selectors'
import {getStatus} from 'src/resources/selectors'

// Types
import {AppState, Bucket, ResourceType, RemoteDataState} from 'src/types'

type ReduxProps = ConnectedProps<typeof connector>
export type Props = ReduxProps

export interface BucketContextType {
  loading: RemoteDataState
  buckets: Bucket[]
  selectedBucketName: string
  setSelectedBucketName: (bucketName: string) => void
}

export const DEFAULT_CONTEXT: BucketContextType = {
  loading: RemoteDataState.NotStarted,
  buckets: [],
  selectedBucketName: '',
  setSelectedBucketName: () => {},
}

export const BucketContext = React.createContext<BucketContextType>(
  DEFAULT_CONTEXT
)

let GLOBAL_LOADING = false

const lockAndLoad = async grabber => {
  GLOBAL_LOADING = true
  await grabber()
  GLOBAL_LOADING = false
}

export const BucketProvider: FC<Props> = ({
  loading,
  getBuckets,
  buckets,
  children,
}) => {
  const [selectedBucketName, setSelectedBucketName] = useState<string>('')

  useEffect(() => {
    if (loading !== RemoteDataState.NotStarted) {
      return
    }

    if (GLOBAL_LOADING) {
      return
    }

    lockAndLoad(getBuckets)
  }, [loading, getBuckets])

  useEffect(() => {
    if (buckets.length && !selectedBucketName) {
      setSelectedBucketName(buckets[0].name)
    }
  }, [buckets, selectedBucketName])

  const setBucketName = useCallback((bucketName: string) => {
    setSelectedBucketName(bucketName)
  }, [])

  return (
    <BucketContext.Provider
      value={{
        loading,
        buckets,
        selectedBucketName,
        setSelectedBucketName: setBucketName,
      }}
    >
      {children}
    </BucketContext.Provider>
  )
}

const mstp = (state: AppState) => {
  const buckets = getSortedBuckets(state)
  const loading = getStatus(state, ResourceType.Buckets)

  return {
    loading,
    buckets,
  }
}

const mdtp = {
  getBuckets: getBuckets,
}

const connector = connect(mstp, mdtp)

export default connector(BucketProvider)
