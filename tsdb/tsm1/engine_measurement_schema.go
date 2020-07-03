package tsm1

import (
	"bytes"
	"context"
	"sort"
	"strings"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/kit/tracing"
	"github.com/influxdata/influxdb/v2/models"
	"github.com/influxdata/influxdb/v2/tsdb"
	"github.com/influxdata/influxdb/v2/tsdb/cursors"
	"github.com/influxdata/influxdb/v2/tsdb/seriesfile"
	"github.com/influxdata/influxql"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// MeasurementNames returns an iterator which enumerates the measurements for the given
// bucket and limited to the time range [start, end].
//
// MeasurementNames will always return a StringIterator if there is no error.
//
// If the context is canceled before MeasurementNames has finished processing, a non-nil
// error will be returned along with statistics for the already scanned data.
func (e *Engine) MeasurementNames(ctx context.Context, orgID, bucketID influxdb.ID, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.Finish()

	return e.tagValuesFast(ctx, orgID, bucketID, nil, models.MeasurementTagKeyBytes, start, end, predicate)
}

// tagValuesCheckInterval represents the period at which tagValuesFast function
// will check for a canceled context. Specifically after every n series
// scanned, the query context will be checked for cancellation, and if canceled,
// the calls will immediately return.
const tagValuesCheckInterval = 64

func (e *Engine) tagValuesFast(ctx context.Context, orgID, bucketID influxdb.ID, measurement, tagKeyBytes []byte, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error) {
	if err := ValidateTagPredicate(predicate); err != nil {
		return nil, err
	}

	orgBucket := tsdb.EncodeName(orgID, bucketID)

	var files []TSMFile
	defer func() {
		for _, f := range files {
			f.Unref()
		}
	}()
	var iters []*TimeRangeIterator

	// TODO(edd): we need to clean up how we're encoding the prefix so that we
	// don't have to remember to get it right everywhere we need to touch TSM data.
	orgBucketEsc := models.EscapeMeasurement(orgBucket[:])

	tsmKeyPrefix := orgBucketEsc
	if len(measurement) > 0 {
		// append the measurement tag key to the prefix
		mt := models.Tags{models.NewTag(models.MeasurementTagKeyBytes, measurement)}
		tsmKeyPrefix = mt.AppendHashKey(tsmKeyPrefix)
		tsmKeyPrefix = append(tsmKeyPrefix, ',')
	}

	var canceled bool

	e.FileStore.ForEachFile(func(f TSMFile) bool {
		// Check the context before accessing each tsm file
		select {
		case <-ctx.Done():
			canceled = true
			return false
		default:
		}
		if f.OverlapsTimeRange(start, end) && f.OverlapsKeyPrefixRange(tsmKeyPrefix, tsmKeyPrefix) {
			f.Ref()
			files = append(files, f)
			iters = append(iters, f.TimeRangeIterator(tsmKeyPrefix, start, end))
		}
		return true
	})

	// fetch distinct values for tag key
	itr, err := e.index.TagValueIterator(orgBucket[:], tagKeyBytes)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	var stats cursors.CursorStats

	if canceled {
		stats = statsFromIters(stats, iters)
		return cursors.NewStringSliceIteratorWithStats(nil, stats), ctx.Err()
	}

	var (
		vals        = make([]string, 0, 128)
		scannedKeys = 0
	)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		defer func() {
			span.LogFields(
				log.Int("files_count", len(files)),
				log.Int("scanned_keys_count", scannedKeys),
				log.Int("values_count", len(vals)),
			)
		}()
	}

	// reusable buffers
	var (
		tags   models.Tags
		keyBuf []byte
		sfkey  []byte
		ts     cursors.TimestampArray
		tagKey = string(tagKeyBytes)
	)

	for i := 0; ; i++ {
		// to keep cache scans fast, check context every 'cancelCheckInterval' iterations
		if i%tagValuesCheckInterval == 0 {
			select {
			case <-ctx.Done():
				stats = statsFromIters(stats, iters)
				return cursors.NewStringSliceIteratorWithStats(nil, stats), ctx.Err()
			default:
			}
		}

		val, err := itr.Next()
		if err != nil {
			stats = statsFromIters(stats, iters)
			return cursors.NewStringSliceIteratorWithStats(nil, stats), err
		} else if len(val) == 0 {
			break
		}

		// <tagKey> = val
		var expr influxql.Expr = &influxql.BinaryExpr{
			LHS: &influxql.VarRef{Val: tagKey, Type: influxql.Tag},
			Op:  influxql.EQ,
			RHS: &influxql.StringLiteral{Val: string(val)},
		}

		if predicate != nil {
			// <tagKey> = val AND (expr)
			expr = &influxql.BinaryExpr{
				LHS: expr,
				Op:  influxql.AND,
				RHS: &influxql.ParenExpr{
					Expr: predicate,
				},
			}
		}

		if err := func() error {
			sitr, err := e.index.MeasurementSeriesByExprIterator(orgBucket[:], expr)
			if err != nil {
				return err
			}
			defer sitr.Close()

			for {
				elem, err := sitr.Next()
				if err != nil {
					return err
				} else if elem.SeriesID.IsZero() {
					return nil
				}

				scannedKeys++

				seriesKey := e.sfile.SeriesKey(elem.SeriesID)
				if len(seriesKey) == 0 {
					continue
				}

				_, tags = seriesfile.ParseSeriesKeyInto(seriesKey, tags[:0])
				if len(tags) < 2 {
					// must contain at least models.MeasurementTagKey and models.FieldTagKey
					continue
				}

				// last value is guaranteed to be field
				fieldVal := tags[len(tags)-1].Value

				// orgBucketEsc is already escaped, so no need to use models.AppendMakeKey, which
				// unescapes and escapes the value again. The degenerate case is if the orgBucketEsc
				// has escaped values, causing two allocations per key
				keyBuf = append(keyBuf[:0], orgBucketEsc...)
				keyBuf = tags.AppendHashKey(keyBuf)
				sfkey = AppendSeriesFieldKeyBytes(sfkey[:0], keyBuf, fieldVal)

				ts.Timestamps = e.Cache.AppendTimestamps(sfkey, ts.Timestamps[:0])
				if ts.Len() > 0 {
					sort.Sort(&ts)

					stats.ScannedValues += ts.Len()
					stats.ScannedBytes += ts.Len() * 8 // sizeof timestamp

					if ts.Contains(start, end) {
						vals = append(vals, string(val))
						return nil
					}
				}

				for _, iter := range iters {
					if exact, _ := iter.Seek(sfkey); !exact {
						continue
					}

					if iter.HasData() {
						vals = append(vals, string(val))
						return nil
					}
				}
			}
		}(); err != nil {
			stats = statsFromIters(stats, iters)
			return cursors.NewStringSliceIteratorWithStats(nil, stats), err
		}
	}

	sort.Strings(vals)
	stats = statsFromIters(stats, iters)
	return cursors.NewStringSliceIteratorWithStats(vals, stats), err
}

// MeasurementTagValues returns an iterator which enumerates the tag values for the given
// bucket, measurement and tag key, filtered using the optional the predicate and limited to the
// time range [start, end].
//
// MeasurementTagValues will always return a StringIterator if there is no error.
//
// If the context is canceled before TagValues has finished processing, a non-nil
// error will be returned along with statistics for the already scanned data.
func (e *Engine) MeasurementTagValues(ctx context.Context, orgID, bucketID influxdb.ID, measurement, tagKey string, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error) {
	predicate = AddMeasurementToExpr(measurement, predicate)

	return e.tagValuesFast(ctx, orgID, bucketID, []byte(measurement), []byte(tagKey), start, end, predicate)

	if predicate == nil {
		return e.tagValuesNoPredicate(ctx, orgID, bucketID, []byte(measurement), []byte(tagKey), start, end)
	}

	predicate = AddMeasurementToExpr(measurement, predicate)

	return e.tagValuesPredicate(ctx, orgID, bucketID, []byte(measurement), []byte(tagKey), start, end, predicate)

}

// MeasurementTagKeys returns an iterator which enumerates the tag keys for the given
// bucket and measurement, filtered using the optional the predicate and limited to the
//// time range [start, end].
//
// MeasurementTagKeys will always return a StringIterator if there is no error.
//
// If the context is canceled before MeasurementTagKeys has finished processing, a non-nil
// error will be returned along with statistics for the already scanned data.
func (e *Engine) MeasurementTagKeys(ctx context.Context, orgID, bucketID influxdb.ID, measurement string, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error) {
	if predicate == nil {
		return e.tagKeysNoPredicate(ctx, orgID, bucketID, []byte(measurement), start, end)
	}

	predicate = AddMeasurementToExpr(measurement, predicate)

	return e.tagKeysPredicate(ctx, orgID, bucketID, []byte(measurement), start, end, predicate)
}

// MeasurementFields returns an iterator which enumerates the field schema for the given
// bucket and measurement, filtered using the optional the predicate and limited to the
//// time range [start, end].
//
// MeasurementFields will always return a MeasurementFieldsIterator if there is no error.
//
// If the context is canceled before MeasurementFields has finished processing, a non-nil
// error will be returned along with statistics for the already scanned data.
func (e *Engine) MeasurementFields(ctx context.Context, orgID, bucketID influxdb.ID, measurement string, start, end int64, predicate influxql.Expr) (cursors.MeasurementFieldsIterator, error) {
	predicate = AddMeasurementToExpr(measurement, predicate)

	return e.fieldsFast(ctx, orgID, bucketID, []byte(measurement), start, end, predicate)

	if predicate == nil {
		return e.fieldsNoPredicate(ctx, orgID, bucketID, []byte(measurement), start, end)
	}

	predicate = AddMeasurementToExpr(measurement, predicate)

	return e.fieldsPredicate(ctx, orgID, bucketID, []byte(measurement), start, end, predicate)
}

type fieldTypeTime struct {
	key []byte
	typ cursors.FieldType
	max int64
}

func (e *Engine) fieldsFast(ctx context.Context, orgID, bucketID influxdb.ID, measurement []byte, start, end int64, predicate influxql.Expr) (cursors.MeasurementFieldsIterator, error) {
	if err := ValidateTagPredicate(predicate); err != nil {
		return nil, err
	}

	orgBucket := tsdb.EncodeName(orgID, bucketID)

	var files []TSMFile
	defer func() {
		for _, f := range files {
			f.Unref()
		}
	}()
	var iters []*TimeRangeMaxTimeIterator

	// TODO(edd): we need to clean up how we're encoding the prefix so that we
	// don't have to remember to get it right everywhere we need to touch TSM data.
	orgBucketEsc := models.EscapeMeasurement(orgBucket[:])

	tsmKeyPrefix := orgBucketEsc
	if len(measurement) > 0 {
		// append the measurement tag key to the prefix
		mt := models.Tags{models.NewTag(models.MeasurementTagKeyBytes, measurement)}
		tsmKeyPrefix = mt.AppendHashKey(tsmKeyPrefix)
		tsmKeyPrefix = append(tsmKeyPrefix, ',')
	}

	var canceled bool

	e.FileStore.ForEachFile(func(f TSMFile) bool {
		// Check the context before accessing each tsm file
		select {
		case <-ctx.Done():
			canceled = true
			return false
		default:
		}
		if f.OverlapsTimeRange(start, end) && f.OverlapsKeyPrefixRange(tsmKeyPrefix, tsmKeyPrefix) {
			f.Ref()
			files = append(files, f)
			iters = append(iters, f.TimeRangeMaxTimeIterator(tsmKeyPrefix, start, end))
		}
		return true
	})

	// fetch distinct values for field, which may be a superset of the measurement
	itr, err := e.index.TagValueIterator(orgBucket[:], models.FieldKeyTagKeyBytes)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	var stats cursors.CursorStats

	if canceled {
		stats = statsFromTimeRangeMaxTimeIters(stats, iters)
		return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), ctx.Err()
	}

	var (
		fieldTypes  = make([]fieldTypeTime, 0, 128)
		scannedKeys = 0
	)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		defer func() {
			span.LogFields(
				log.Int("files_count", len(files)),
				log.Int("scanned_keys_count", scannedKeys),
				log.Int("values_count", len(fieldTypes)),
			)
		}()
	}

	// reusable buffers
	var (
		tags   models.Tags
		keyBuf []byte
		sfkey  []byte
		ts     cursors.TimestampArray
		tagKey = models.FieldKeyTagKey
	)

	for i := 0; ; i++ {
		// to keep cache scans fast, check context every 'cancelCheckInterval' iterations
		if i%tagValuesCheckInterval == 0 {
			select {
			case <-ctx.Done():
				stats = statsFromTimeRangeMaxTimeIters(stats, iters)
				return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), ctx.Err()
			default:
			}
		}

		val, err := itr.Next()
		if err != nil {
			stats = statsFromTimeRangeMaxTimeIters(stats, iters)
			return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), err
		} else if len(val) == 0 {
			break
		}

		// <tagKey> = val
		var expr influxql.Expr = &influxql.BinaryExpr{
			LHS: &influxql.VarRef{Val: tagKey, Type: influxql.Tag},
			Op:  influxql.EQ,
			RHS: &influxql.StringLiteral{Val: string(val)},
		}

		if predicate != nil {
			// <tagKey> = val AND (expr)
			expr = &influxql.BinaryExpr{
				LHS: expr,
				Op:  influxql.AND,
				RHS: &influxql.ParenExpr{
					Expr: predicate,
				},
			}
		}

		if err := func() error {
			sitr, err := e.index.MeasurementSeriesByExprIterator(orgBucket[:], expr)
			if err != nil {
				return err
			}
			defer sitr.Close()

			for {
				elem, err := sitr.Next()
				if err != nil {
					return err
				} else if elem.SeriesID.IsZero() {
					return nil
				}

				scannedKeys++

				seriesKey := e.sfile.SeriesKey(elem.SeriesID)
				if len(seriesKey) == 0 {
					continue
				}

				_, tags = seriesfile.ParseSeriesKeyInto(seriesKey, tags[:0])
				if len(tags) < 2 {
					// must contain at least models.MeasurementTagKey and models.FieldTagKey
					continue
				}

				// last value is guaranteed to be field
				fieldVal := tags[len(tags)-1].Value

				// orgBucketEsc is already escaped, so no need to use models.AppendMakeKey, which
				// unescapes and escapes the value again. The degenerate case is if the orgBucketEsc
				// has escaped values, causing two allocations per key
				keyBuf = append(keyBuf[:0], orgBucketEsc...)
				keyBuf = tags.AppendHashKey(keyBuf)
				sfkey = AppendSeriesFieldKeyBytes(sfkey[:0], keyBuf, fieldVal)

				cur := fieldTypeTime{key: fieldVal, max: InvalidMinNanoTime}

				ts.Timestamps = e.Cache.AppendTimestamps(sfkey, ts.Timestamps[:0])
				if ts.Len() > 0 {
					sort.Sort(&ts)

					stats.ScannedValues += ts.Len()
					stats.ScannedBytes += ts.Len() * 8 // sizeof timestamp

					if ts.Contains(start, end) {
						max := ts.MaxTime()
						if max > cur.max {
							cur.max = max
							cur.typ = BlockTypeToFieldType(e.Cache.BlockType(sfkey))
						}
					}
				}

				for _, iter := range iters {
					if exact, _ := iter.Seek(sfkey); !exact {
						continue
					}

					max := iter.MaxTime()
					if max > cur.max {
						cur.max = max
						cur.typ = BlockTypeToFieldType(iter.Type())
					}
				}

				if cur.max != InvalidMinNanoTime {
					fieldTypes = append(fieldTypes, cur)
					return nil
				}
			}
		}(); err != nil {
			stats = statsFromTimeRangeMaxTimeIters(stats, iters)
			return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), err
		}
	}

	vals := make([]cursors.MeasurementField, 0, len(fieldTypes))
	for i := range fieldTypes {
		val := &fieldTypes[i]
		vals = append(vals, cursors.MeasurementField{Key: string(val.key), Type: val.typ, Timestamp: val.max})
	}

	stats = statsFromTimeRangeMaxTimeIters(stats, iters)
	return cursors.NewMeasurementFieldsSliceIteratorWithStats([]cursors.MeasurementFields{{Fields: vals}}, stats), nil
}

func (e *Engine) fieldsPredicate(ctx context.Context, orgID influxdb.ID, bucketID influxdb.ID, measurement []byte, start int64, end int64, predicate influxql.Expr) (cursors.MeasurementFieldsIterator, error) {
	if err := ValidateTagPredicate(predicate); err != nil {
		return nil, err
	}

	orgBucket := tsdb.EncodeName(orgID, bucketID)

	keys, err := e.findCandidateKeys(ctx, orgBucket[:], predicate)
	if err != nil {
		return cursors.EmptyMeasurementFieldsIterator, err
	}

	if len(keys) == 0 {
		return cursors.EmptyMeasurementFieldsIterator, nil
	}

	var files []TSMFile
	defer func() {
		for _, f := range files {
			f.Unref()
		}
	}()
	var iters []*TimeRangeMaxTimeIterator

	// TODO(edd): we need to clean up how we're encoding the prefix so that we
	// don't have to remember to get it right everywhere we need to touch TSM data.
	orgBucketEsc := models.EscapeMeasurement(orgBucket[:])

	mt := models.Tags{models.NewTag(models.MeasurementTagKeyBytes, measurement)}
	tsmKeyPrefix := mt.AppendHashKey(orgBucketEsc)
	tsmKeyPrefix = append(tsmKeyPrefix, ',')

	var canceled bool

	e.FileStore.ForEachFile(func(f TSMFile) bool {
		// Check the context before accessing each tsm file
		select {
		case <-ctx.Done():
			canceled = true
			return false
		default:
		}
		if f.OverlapsTimeRange(start, end) && f.OverlapsKeyPrefixRange(tsmKeyPrefix, tsmKeyPrefix) {
			f.Ref()
			files = append(files, f)
			iters = append(iters, f.TimeRangeMaxTimeIterator(tsmKeyPrefix, start, end))
		}
		return true
	})

	var stats cursors.CursorStats

	if canceled {
		stats = statsFromTimeRangeMaxTimeIters(stats, iters)
		return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), ctx.Err()
	}

	tsmValues := make(map[string]fieldTypeTime)

	// reusable buffers
	var (
		tags   models.Tags
		keybuf []byte
		sfkey  []byte
		ts     cursors.TimestampArray
	)

	for i := range keys {
		// to keep cache scans fast, check context every 'cancelCheckInterval' iteratons
		if i%cancelCheckInterval == 0 {
			select {
			case <-ctx.Done():
				stats = statsFromTimeRangeMaxTimeIters(stats, iters)
				return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), ctx.Err()
			default:
			}
		}

		_, tags = seriesfile.ParseSeriesKeyInto(keys[i], tags[:0])
		fieldKey := tags.Get(models.FieldKeyTagKeyBytes)
		keybuf = models.AppendMakeKey(keybuf[:0], orgBucketEsc, tags)
		sfkey = AppendSeriesFieldKeyBytes(sfkey[:0], keybuf, fieldKey)

		cur := fieldTypeTime{max: InvalidMinNanoTime}

		ts.Timestamps = e.Cache.AppendTimestamps(sfkey, ts.Timestamps[:0])
		if ts.Len() > 0 {
			sort.Sort(&ts)

			stats.ScannedValues += ts.Len()
			stats.ScannedBytes += ts.Len() * 8 // sizeof timestamp

			if ts.Contains(start, end) {
				max := ts.MaxTime()
				if max > cur.max {
					cur.max = max
					cur.typ = BlockTypeToFieldType(e.Cache.BlockType(sfkey))
				}
			}
		}

		for _, iter := range iters {
			if exact, _ := iter.Seek(sfkey); !exact {
				continue
			}

			max := iter.MaxTime()
			if max > cur.max {
				cur.max = max
				cur.typ = BlockTypeToFieldType(iter.Type())
			}
		}

		if cur.max != InvalidMinNanoTime {
			tsmValues[string(fieldKey)] = cur
		}
	}

	vals := make([]cursors.MeasurementField, 0, len(tsmValues))
	for key, val := range tsmValues {
		vals = append(vals, cursors.MeasurementField{Key: key, Type: val.typ, Timestamp: val.max})
	}

	return cursors.NewMeasurementFieldsSliceIteratorWithStats([]cursors.MeasurementFields{{Fields: vals}}, stats), nil
}

func (e *Engine) fieldsNoPredicate(ctx context.Context, orgID influxdb.ID, bucketID influxdb.ID, measurement []byte, start int64, end int64) (cursors.MeasurementFieldsIterator, error) {
	tsmValues := make(map[string]fieldTypeTime)
	orgBucket := tsdb.EncodeName(orgID, bucketID)

	// TODO(edd): we need to clean up how we're encoding the prefix so that we
	// don't have to remember to get it right everywhere we need to touch TSM data.
	orgBucketEsc := models.EscapeMeasurement(orgBucket[:])

	mt := models.Tags{models.NewTag(models.MeasurementTagKeyBytes, measurement)}
	tsmKeyPrefix := mt.AppendHashKey(orgBucketEsc)
	tsmKeyPrefix = append(tsmKeyPrefix, ',')

	var stats cursors.CursorStats
	var canceled bool

	e.FileStore.ForEachFile(func(f TSMFile) bool {
		// Check the context before touching each tsm file
		select {
		case <-ctx.Done():
			canceled = true
			return false
		default:
		}
		if f.OverlapsTimeRange(start, end) && f.OverlapsKeyPrefixRange(tsmKeyPrefix, tsmKeyPrefix) {
			// TODO(sgc): create f.TimeRangeIterator(minKey, maxKey, start, end)
			iter := f.TimeRangeMaxTimeIterator(tsmKeyPrefix, start, end)
			for i := 0; iter.Next(); i++ {
				sfkey := iter.Key()
				if !bytes.HasPrefix(sfkey, tsmKeyPrefix) {
					// end of prefix
					break
				}

				max := iter.MaxTime()
				if max == InvalidMinNanoTime {
					continue
				}

				_, fieldKey := SeriesAndFieldFromCompositeKey(sfkey)
				v, ok := tsmValues[string(fieldKey)]
				if !ok || v.max < max {
					tsmValues[string(fieldKey)] = fieldTypeTime{
						typ: BlockTypeToFieldType(iter.Type()),
						max: max,
					}
				}
			}
			stats.Add(iter.Stats())
		}
		return true
	})

	if canceled {
		return cursors.NewMeasurementFieldsSliceIteratorWithStats(nil, stats), ctx.Err()
	}

	var ts cursors.TimestampArray

	// With performance in mind, we explicitly do not check the context
	// while scanning the entries in the cache.
	tsmKeyPrefixStr := string(tsmKeyPrefix)
	_ = e.Cache.ApplyEntryFn(func(sfkey string, entry *entry) error {
		if !strings.HasPrefix(sfkey, tsmKeyPrefixStr) {
			return nil
		}

		ts.Timestamps = entry.AppendTimestamps(ts.Timestamps[:0])
		if ts.Len() == 0 {
			return nil
		}

		sort.Sort(&ts)

		stats.ScannedValues += ts.Len()
		stats.ScannedBytes += ts.Len() * 8 // sizeof timestamp

		if !ts.Contains(start, end) {
			return nil
		}

		max := ts.MaxTime()

		// TODO(edd): consider the []byte() conversion here.
		_, fieldKey := SeriesAndFieldFromCompositeKey([]byte(sfkey))
		v, ok := tsmValues[string(fieldKey)]
		if !ok || v.max < max {
			tsmValues[string(fieldKey)] = fieldTypeTime{
				typ: BlockTypeToFieldType(entry.BlockType()),
				max: max,
			}
		}

		return nil
	})

	vals := make([]cursors.MeasurementField, 0, len(tsmValues))
	for key, val := range tsmValues {
		vals = append(vals, cursors.MeasurementField{Key: key, Type: val.typ, Timestamp: val.max})
	}

	return cursors.NewMeasurementFieldsSliceIteratorWithStats([]cursors.MeasurementFields{{Fields: vals}}, stats), nil
}

func AddMeasurementToExpr(measurement string, base influxql.Expr) influxql.Expr {
	// \x00 = '<measurement>'
	expr := &influxql.BinaryExpr{
		LHS: &influxql.VarRef{
			Val:  models.MeasurementTagKey,
			Type: influxql.Tag,
		},
		Op: influxql.EQ,
		RHS: &influxql.StringLiteral{
			Val: measurement,
		},
	}

	if base != nil {
		// \x00 = '<measurement>' AND (base)
		expr = &influxql.BinaryExpr{
			LHS: expr,
			Op:  influxql.AND,
			RHS: &influxql.ParenExpr{
				Expr: base,
			},
		}
	}

	return expr
}

func statsFromTimeRangeMaxTimeIters(stats cursors.CursorStats, iters []*TimeRangeMaxTimeIterator) cursors.CursorStats {
	for _, iter := range iters {
		stats.Add(iter.Stats())
	}
	return stats
}
