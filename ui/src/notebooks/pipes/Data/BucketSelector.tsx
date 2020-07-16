// Libraries
import React, {FC, useContext} from 'react'

// Components
import {
  DapperScrollbars,
  TechnoSpinner,
  ComponentSize,
  RemoteDataState,
} from '@influxdata/clockface'
import SelectorListItem from 'src/notebooks/pipes/Data/SelectorListItem'
import {BucketContext} from 'src/notebooks/context/buckets'

const BucketSelector: FC = () => {
  const {
    buckets,
    loading,
    selectedBucketName,
    setSelectedBucketName,
  } = useContext(BucketContext)

  const handleItemClick = (bucket: Bucket): void => {
    const bucketName = bucket.name

    setSelectedBucketName(bucketName)
  }

  let body

  if (loading === RemoteDataState.Loading) {
    body = (
      <div className="data-source--list__empty">
        <TechnoSpinner strokeWidth={ComponentSize.Small} diameterPixels={32} />
      </div>
    )
  }

  if (loading === RemoteDataState.Error) {
    body = (
      <div className="data-source--list__empty">
        <p>Could not fetch Buckets</p>
      </div>
    )
  }

  if (loading === RemoteDataState.Done && selectedBucketName) {
    body = (
      <DapperScrollbars className="data-source--list">
        {buckets.map(bucket => (
          <SelectorListItem
            key={bucket.name}
            value={bucket}
            onClick={handleItemClick}
            selected={bucket.name === selectedBucketName}
            text={bucket.name}
          />
        ))}
      </DapperScrollbars>
    )
  }

  console.log(selectedBucketName)

  return (
    <div className="data-source--block">
      <div className="data-source--block-title">Bucket</div>
      {body}
    </div>
  )
}

export default BucketSelector
