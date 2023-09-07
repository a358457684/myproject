package redis

import (
	"common/config"
	"common/log"
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/suiyunonghen/DxCommonLib"
	"time"
)

var (
	ErrLockerTimeout = errors.New("执行超时")
)

func init() {
	if config.Data.Redis == nil {
		log.Warn("跳过Redis初始化，读取Redis配置失败")
		return
	}
	err := initRedis(config.Data.Redis)
	if err != nil {
		log.WithError(err).Error("Redis初始化失败")
	} else {
		log.Info("Redis初始化成功")
	}
	//重置一下时间轮，精度设置到50毫秒
	DxCommonLib.ReSetDefaultTimeWheel(time.Millisecond*50, 7200, nil)
}

func Command(ctx context.Context) *redis.CommandsInfoCmd {
	cmd := Client.Command(ctx)
	logError(cmd)
	return cmd
}

func IsRedisNil(err error) bool {
	return redis.Nil == err
}

func ClientGetName(ctx context.Context) *redis.StringCmd {
	cmd := Client.ClientGetName(ctx)
	logError(cmd)
	return cmd
}
func Echo(ctx context.Context, message interface{}) *redis.StringCmd {
	cmd := Client.Echo(ctx, message)
	logError(cmd)
	return cmd
}
func Ping(ctx context.Context) *redis.StatusCmd {
	cmd := Client.Ping(ctx)
	logError(cmd)
	return cmd
}
func Quit(ctx context.Context) *redis.StatusCmd {
	cmd := Client.Quit(ctx)
	logError(cmd)
	return cmd
}
func Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.Del(ctx, keys...)
	logError(cmd)
	return cmd
}
func Unlink(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.Unlink(ctx, keys...)
	logError(cmd)
	return cmd
}
func Dump(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.Dump(ctx, key)
	logError(cmd)
	return cmd
}
func Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.Exists(ctx, keys...)
	logError(cmd)
	return cmd
}
func Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := Client.Expire(ctx, key, expiration)
	logError(cmd)
	return cmd
}
func ExpireAt(ctx context.Context, key string, tm time.Time) *redis.BoolCmd {
	cmd := Client.ExpireAt(ctx, key, tm)
	logError(cmd)
	return cmd
}
func Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := Client.Keys(ctx, pattern)
	logError(cmd)
	return cmd
}
func Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) *redis.StatusCmd {
	cmd := Client.Migrate(ctx, host, port, key, db, timeout)
	logError(cmd)
	return cmd
}
func Move(ctx context.Context, key string, db int) *redis.BoolCmd {
	cmd := Client.Move(ctx, key, db)
	logError(cmd)
	return cmd
}
func ObjectRefCount(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.ObjectRefCount(ctx, key)
	logError(cmd)
	return cmd
}
func ObjectEncoding(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.ObjectEncoding(ctx, key)
	logError(cmd)
	return cmd
}
func ObjectIdleTime(ctx context.Context, key string) *redis.DurationCmd {
	cmd := Client.ObjectIdleTime(ctx, key)
	logError(cmd)
	return cmd
}
func Persist(ctx context.Context, key string) *redis.BoolCmd {
	cmd := Client.Persist(ctx, key)
	logError(cmd)
	return cmd
}
func PExpire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := Client.PExpire(ctx, key, expiration)
	logError(cmd)
	return cmd
}
func PExpireAt(ctx context.Context, key string, tm time.Time) *redis.BoolCmd {
	cmd := Client.PExpireAt(ctx, key, tm)
	logError(cmd)
	return cmd
}
func PTTL(ctx context.Context, key string) *redis.DurationCmd {
	cmd := Client.PTTL(ctx, key)
	logError(cmd)
	return cmd
}
func RandomKey(ctx context.Context) *redis.StringCmd {
	cmd := Client.RandomKey(ctx)
	logError(cmd)
	return cmd
}
func Rename(ctx context.Context, key, newkey string) *redis.StatusCmd {
	cmd := Client.Rename(ctx, key, newkey)
	logError(cmd)
	return cmd
}
func RenameNX(ctx context.Context, key, newkey string) *redis.BoolCmd {
	cmd := Client.RenameNX(ctx, key, newkey)
	logError(cmd)
	return cmd
}
func Restore(ctx context.Context, key string, ttl time.Duration, value string) *redis.StatusCmd {
	cmd := Client.Restore(ctx, key, ttl, value)
	logError(cmd)
	return cmd
}
func RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) *redis.StatusCmd {
	cmd := Client.RestoreReplace(ctx, key, ttl, value)
	logError(cmd)
	return cmd
}
func Sort(ctx context.Context, key string, sort *redis.Sort) *redis.StringSliceCmd {
	cmd := Client.Sort(ctx, key, sort)
	logError(cmd)
	return cmd
}
func SortStore(ctx context.Context, key, store string, sort *redis.Sort) *redis.IntCmd {
	cmd := Client.SortStore(ctx, key, store, sort)
	logError(cmd)
	return cmd
}
func SortInterfaces(ctx context.Context, key string, sort *redis.Sort) *redis.SliceCmd {
	cmd := Client.SortInterfaces(ctx, key, sort)
	logError(cmd)
	return cmd
}
func Touch(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.Touch(ctx, keys...)
	logError(cmd)
	return cmd
}
func TTL(ctx context.Context, key string) *redis.DurationCmd {
	cmd := Client.TTL(ctx, key)
	logError(cmd)
	return cmd
}
func Type(ctx context.Context, key string) *redis.StatusCmd {
	cmd := Client.Type(ctx, key)
	logError(cmd)
	return cmd
}
func Append(ctx context.Context, key, value string) *redis.IntCmd {
	cmd := Client.Append(ctx, key, value)
	logError(cmd)
	return cmd
}
func Decr(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.Decr(ctx, key)
	logError(cmd)
	return cmd
}
func DecrBy(ctx context.Context, key string, decrement int64) *redis.IntCmd {
	cmd := Client.DecrBy(ctx, key, decrement)
	logError(cmd)
	return cmd
}
func Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.Get(ctx, key)
	logError(cmd)
	return cmd
}
func GetRange(ctx context.Context, key string, start, end int64) *redis.StringCmd {
	cmd := Client.GetRange(ctx, key, start, end)
	logError(cmd)
	return cmd
}
func GetSet(ctx context.Context, key string, value interface{}) *redis.StringCmd {
	cmd := Client.GetSet(ctx, key, value)
	logError(cmd)
	return cmd
}
func Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.Incr(ctx, key)
	logError(cmd)
	return cmd
}
func IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	cmd := Client.IncrBy(ctx, key, value)
	logError(cmd)
	return cmd
}
func IncrByFloat(ctx context.Context, key string, value float64) *redis.FloatCmd {
	cmd := Client.IncrByFloat(ctx, key, value)
	logError(cmd)
	return cmd
}
func MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	cmd := Client.MGet(ctx, keys...)
	logError(cmd)
	return cmd
}
func MSet(ctx context.Context, values ...interface{}) *redis.StatusCmd {
	cmd := Client.MSet(ctx, values...)
	logError(cmd)
	return cmd
}
func MSetNX(ctx context.Context, values ...interface{}) *redis.BoolCmd {
	cmd := Client.MSetNX(ctx, values...)
	logError(cmd)
	return cmd
}
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := Client.Set(ctx, key, value, expiration)
	logError(cmd)
	return cmd
}
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	cmd := Client.SetNX(ctx, key, value, expiration)
	logError(cmd)
	return cmd
}
func SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	cmd := Client.SetXX(ctx, key, value, expiration)
	logError(cmd)
	return cmd
}
func SetRange(ctx context.Context, key string, offset int64, value string) *redis.IntCmd {
	cmd := Client.SetRange(ctx, key, offset, value)
	logError(cmd)
	return cmd
}
func StrLen(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.StrLen(ctx, key)
	logError(cmd)
	return cmd
}

func GetBit(ctx context.Context, key string, offset int64) *redis.IntCmd {
	cmd := Client.GetBit(ctx, key, offset)
	logError(cmd)
	return cmd
}
func SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd {
	cmd := Client.SetBit(ctx, key, offset, value)
	logError(cmd)
	return cmd
}
func BitCount(ctx context.Context, key string, bitCount *redis.BitCount) *redis.IntCmd {
	cmd := Client.BitCount(ctx, key, bitCount)
	logError(cmd)
	return cmd
}
func BitOpAnd(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	cmd := Client.BitOpAnd(ctx, destKey, keys...)
	logError(cmd)
	return cmd
}
func BitOpOr(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	cmd := Client.BitOpOr(ctx, destKey, keys...)
	logError(cmd)
	return cmd
}
func BitOpXor(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	cmd := Client.BitOpXor(ctx, destKey, keys...)
	logError(cmd)
	return cmd
}
func BitOpNot(ctx context.Context, destKey string, key string) *redis.IntCmd {
	cmd := Client.BitOpNot(ctx, destKey, key)
	logError(cmd)
	return cmd
}
func BitPos(ctx context.Context, key string, bit int64, pos ...int64) *redis.IntCmd {
	cmd := Client.BitPos(ctx, key, bit, pos...)
	logError(cmd)
	return cmd
}
func BitField(ctx context.Context, key string, args ...interface{}) *redis.IntSliceCmd {
	cmd := Client.BitField(ctx, key, args...)
	logError(cmd)
	return cmd
}

func Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := Client.Scan(ctx, cursor, match, count)
	logError(cmd)
	return cmd
}
func SScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := Client.SScan(ctx, key, cursor, match, count)
	logError(cmd)
	return cmd
}
func HScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := Client.HScan(ctx, key, cursor, match, count)
	logError(cmd)
	return cmd
}
func ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := Client.ZScan(ctx, key, cursor, match, count)
	logError(cmd)
	return cmd
}

func HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	cmd := Client.HDel(ctx, key, fields...)
	logError(cmd)
	return cmd
}
func HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	cmd := Client.HExists(ctx, key, field)
	logError(cmd)
	return cmd
}
func HGet(ctx context.Context, key, field string) *redis.StringCmd {
	cmd := Client.HGet(ctx, key, field)
	logError(cmd)
	return cmd
}
func HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	cmd := Client.HGetAll(ctx, key)
	logError(cmd)
	return cmd
}
func HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	cmd := Client.HIncrBy(ctx, key, field, incr)
	logError(cmd)
	return cmd
}
func HIncrByFloat(ctx context.Context, key, field string, incr float64) *redis.FloatCmd {
	cmd := Client.HIncrByFloat(ctx, key, field, incr)
	logError(cmd)
	return cmd
}
func HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	cmd := Client.HKeys(ctx, key)
	logError(cmd)
	return cmd
}
func HLen(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.HLen(ctx, key)
	logError(cmd)
	return cmd
}
func HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	cmd := Client.HMGet(ctx, key, fields...)
	logError(cmd)
	return cmd
}
func HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := Client.HSet(ctx, key, values...)
	logError(cmd)
	return cmd
}
func HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	cmd := Client.HMSet(ctx, key, values...)
	logError(cmd)
	return cmd
}
func HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd {
	cmd := Client.HSetNX(ctx, key, field, value)
	logError(cmd)
	return cmd
}
func HVals(ctx context.Context, key string) *redis.StringSliceCmd {
	cmd := Client.HVals(ctx, key)
	logError(cmd)
	return cmd
}

func BLPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	cmd := Client.BLPop(ctx, timeout, keys...)
	logError(cmd)
	return cmd
}
func BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	cmd := Client.BRPop(ctx, timeout, keys...)
	logError(cmd)
	return cmd
}
func BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) *redis.StringCmd {
	cmd := Client.BRPopLPush(ctx, source, destination, timeout)
	logError(cmd)
	return cmd
}
func LIndex(ctx context.Context, key string, index int64) *redis.StringCmd {
	cmd := Client.LIndex(ctx, key, index)
	logError(cmd)
	return cmd
}
func LInsert(ctx context.Context, key, op string, pivot, value interface{}) *redis.IntCmd {
	cmd := Client.LInsert(ctx, key, op, pivot, value)
	logError(cmd)
	return cmd
}
func LInsertBefore(ctx context.Context, key string, pivot, value interface{}) *redis.IntCmd {
	cmd := Client.LInsertBefore(ctx, key, pivot, value)
	logError(cmd)
	return cmd
}
func LInsertAfter(ctx context.Context, key string, pivot, value interface{}) *redis.IntCmd {
	cmd := Client.LInsertAfter(ctx, key, pivot, value)
	logError(cmd)
	return cmd
}
func LLen(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.LLen(ctx, key)
	logError(cmd)
	return cmd
}
func LPop(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.LPop(ctx, key)
	logError(cmd)
	return cmd
}
func LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := Client.LPush(ctx, key, values...)
	logError(cmd)
	return cmd
}
func LPushX(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := Client.LPushX(ctx, key, values...)
	logError(cmd)
	return cmd
}
func LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	cmd := Client.LRange(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd {
	cmd := Client.LRem(ctx, key, count, value)
	logError(cmd)
	return cmd
}
func LSet(ctx context.Context, key string, index int64, value interface{}) *redis.StatusCmd {
	cmd := Client.LSet(ctx, key, index, value)
	logError(cmd)
	return cmd
}
func LTrim(ctx context.Context, key string, start, stop int64) *redis.StatusCmd {
	cmd := Client.LTrim(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func RPop(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.RPop(ctx, key)
	logError(cmd)
	return cmd
}
func RPopLPush(ctx context.Context, source, destination string) *redis.StringCmd {
	cmd := Client.RPopLPush(ctx, source, destination)
	logError(cmd)
	return cmd
}
func RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := Client.RPush(ctx, key, values...)
	logError(cmd)
	return cmd
}
func RPushX(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	cmd := Client.RPushX(ctx, key, values...)
	logError(cmd)
	return cmd
}

func SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := Client.SAdd(ctx, key, members...)
	logError(cmd)
	return cmd
}
func SCard(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.SCard(ctx, key)
	logError(cmd)
	return cmd
}
func SDiff(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	cmd := Client.SDiff(ctx, keys...)
	logError(cmd)
	return cmd
}
func SDiffStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	cmd := Client.SDiffStore(ctx, destination, keys...)
	logError(cmd)
	return cmd
}
func SInter(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	cmd := Client.SInter(ctx, keys...)
	logError(cmd)
	return cmd
}
func SInterStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	cmd := Client.SInterStore(ctx, destination, keys...)
	logError(cmd)
	return cmd
}
func SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	cmd := Client.SIsMember(ctx, key, member)
	logError(cmd)
	return cmd
}
func SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	cmd := Client.SMembers(ctx, key)
	logError(cmd)
	return cmd
}
func SMembersMap(ctx context.Context, key string) *redis.StringStructMapCmd {
	cmd := Client.SMembersMap(ctx, key)
	logError(cmd)
	return cmd
}
func SMove(ctx context.Context, source, destination string, member interface{}) *redis.BoolCmd {
	cmd := Client.SMove(ctx, source, destination, member)
	logError(cmd)
	return cmd
}
func SPop(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.SPop(ctx, key)
	logError(cmd)
	return cmd
}
func SPopN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	cmd := Client.SPopN(ctx, key, count)
	logError(cmd)
	return cmd
}
func SRandMember(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.SRandMember(ctx, key)
	logError(cmd)
	return cmd
}
func SRandMemberN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	cmd := Client.SRandMemberN(ctx, key, count)
	logError(cmd)
	return cmd
}
func SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := Client.SRem(ctx, key, members)
	logError(cmd)
	return cmd
}
func SUnion(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	cmd := Client.SUnion(ctx, keys...)
	logError(cmd)
	return cmd
}
func SUnionStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	cmd := Client.SUnionStore(ctx, destination, keys...)
	logError(cmd)
	return cmd
}

func XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	cmd := Client.XAdd(ctx, a)
	logError(cmd)
	return cmd
}
func XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd {
	cmd := Client.XDel(ctx, stream, ids...)
	logError(cmd)
	return cmd
}
func XLen(ctx context.Context, stream string) *redis.IntCmd {
	cmd := Client.XLen(ctx, stream)
	logError(cmd)
	return cmd
}
func XRange(ctx context.Context, stream, start, stop string) *redis.XMessageSliceCmd {
	cmd := Client.XRange(ctx, stream, start, stop)
	logError(cmd)
	return cmd
}
func XRangeN(ctx context.Context, stream, start, stop string, count int64) *redis.XMessageSliceCmd {
	cmd := Client.XRangeN(ctx, stream, start, stop, count)
	logError(cmd)
	return cmd
}
func XRevRange(ctx context.Context, stream string, start, stop string) *redis.XMessageSliceCmd {
	cmd := Client.XRevRange(ctx, stream, start, stop)
	logError(cmd)
	return cmd
}
func XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) *redis.XMessageSliceCmd {
	cmd := Client.XRevRangeN(ctx, stream, start, stop, count)
	logError(cmd)
	return cmd
}
func XRead(ctx context.Context, a *redis.XReadArgs) *redis.XStreamSliceCmd {
	cmd := Client.XRead(ctx, a)
	logError(cmd)
	return cmd
}
func XReadStreams(ctx context.Context, streams ...string) *redis.XStreamSliceCmd {
	cmd := Client.XReadStreams(ctx, streams...)
	logError(cmd)
	return cmd
}
func XGroupCreate(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	cmd := Client.XGroupCreate(ctx, stream, group, start)
	logError(cmd)
	return cmd
}
func XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	cmd := Client.XGroupCreateMkStream(ctx, stream, group, start)
	logError(cmd)
	return cmd
}
func XGroupSetID(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	cmd := Client.XGroupSetID(ctx, stream, group, start)
	logError(cmd)
	return cmd
}
func XGroupDestroy(ctx context.Context, stream, group string) *redis.IntCmd {
	cmd := Client.XGroupDestroy(ctx, stream, group)
	logError(cmd)
	return cmd
}
func XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd {
	cmd := Client.XGroupDelConsumer(ctx, stream, group, consumer)
	logError(cmd)
	return cmd
}
func XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	cmd := Client.XReadGroup(ctx, a)
	logError(cmd)
	return cmd
}
func XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	cmd := Client.XAck(ctx, stream, group, ids...)
	logError(cmd)
	return cmd
}
func XPending(ctx context.Context, stream, group string) *redis.XPendingCmd {
	cmd := Client.XPending(ctx, stream, group)
	logError(cmd)
	return cmd
}
func XPendingExt(ctx context.Context, a *redis.XPendingExtArgs) *redis.XPendingExtCmd {
	cmd := Client.XPendingExt(ctx, a)
	logError(cmd)
	return cmd
}
func XClaim(ctx context.Context, a *redis.XClaimArgs) *redis.XMessageSliceCmd {
	cmd := Client.XClaim(ctx, a)
	logError(cmd)
	return cmd
}
func XClaimJustID(ctx context.Context, a *redis.XClaimArgs) *redis.StringSliceCmd {
	cmd := Client.XClaimJustID(ctx, a)
	logError(cmd)
	return cmd
}
func XTrim(ctx context.Context, key string, maxLen int64) *redis.IntCmd {
	cmd := Client.XTrim(ctx, key, maxLen)
	logError(cmd)
	return cmd
}
func XTrimApprox(ctx context.Context, key string, maxLen int64) *redis.IntCmd {
	cmd := Client.XTrimApprox(ctx, key, maxLen)
	logError(cmd)
	return cmd
}
func XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd {
	cmd := Client.XInfoGroups(ctx, key)
	logError(cmd)
	return cmd
}

func BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	cmd := Client.BZPopMax(ctx, timeout, keys...)
	logError(cmd)
	return cmd
}
func BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	cmd := Client.BZPopMin(ctx, timeout, keys...)
	logError(cmd)
	return cmd
}
func ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAdd(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZAddNX(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAddNX(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZAddXX(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAddXX(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZAddCh(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAddCh(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZAddNXCh(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAddNXCh(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZAddXXCh(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	cmd := Client.ZAddXXCh(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZIncr(ctx context.Context, key string, member *redis.Z) *redis.FloatCmd {
	cmd := Client.ZIncr(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZIncrNX(ctx context.Context, key string, member *redis.Z) *redis.FloatCmd {
	cmd := Client.ZIncrNX(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZIncrXX(ctx context.Context, key string, member *redis.Z) *redis.FloatCmd {
	cmd := Client.ZIncrXX(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZCard(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.ZCard(ctx, key)
	logError(cmd)
	return cmd
}
func ZCount(ctx context.Context, key, min, max string) *redis.IntCmd {
	cmd := Client.ZCount(ctx, key, min, max)
	logError(cmd)
	return cmd
}
func ZLexCount(ctx context.Context, key, min, max string) *redis.IntCmd {
	cmd := Client.ZLexCount(ctx, key, min, max)
	logError(cmd)
	return cmd
}
func ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	cmd := Client.ZIncrBy(ctx, key, increment, member)
	logError(cmd)
	return cmd
}
func ZInterStore(ctx context.Context, destination string, store *redis.ZStore) *redis.IntCmd {
	cmd := Client.ZInterStore(ctx, destination, store)
	logError(cmd)
	return cmd
}
func ZPopMax(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	cmd := Client.ZPopMax(ctx, key, count...)
	logError(cmd)
	return cmd
}
func ZPopMin(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	cmd := Client.ZPopMin(ctx, key, count...)
	logError(cmd)
	return cmd
}
func ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	cmd := Client.ZRange(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	cmd := Client.ZRangeWithScores(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := Client.ZRangeByScore(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := Client.ZRangeByLex(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	cmd := Client.ZRangeByScoreWithScores(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRank(ctx context.Context, key, member string) *redis.IntCmd {
	cmd := Client.ZRank(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := Client.ZRem(ctx, key, members...)
	logError(cmd)
	return cmd
}
func ZRemRangeByRank(ctx context.Context, key string, start, stop int64) *redis.IntCmd {
	cmd := Client.ZRemRangeByRank(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd {
	cmd := Client.ZRemRangeByScore(ctx, key, min, max)
	logError(cmd)
	return cmd
}
func ZRemRangeByLex(ctx context.Context, key, min, max string) *redis.IntCmd {
	cmd := Client.ZRemRangeByLex(ctx, key, min, max)
	logError(cmd)
	return cmd
}
func ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	cmd := Client.ZRevRange(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	cmd := Client.ZRevRangeWithScores(ctx, key, start, stop)
	logError(cmd)
	return cmd
}
func ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := Client.ZRevRangeByScore(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRevRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := Client.ZRevRangeByLex(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	cmd := Client.ZRevRangeByScoreWithScores(ctx, key, opt)
	logError(cmd)
	return cmd
}
func ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	cmd := Client.ZRevRank(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	cmd := Client.ZScore(ctx, key, member)
	logError(cmd)
	return cmd
}
func ZUnionStore(ctx context.Context, dest string, store *redis.ZStore) *redis.IntCmd {
	cmd := Client.ZUnionStore(ctx, dest, store)
	logError(cmd)
	return cmd
}

func PFAdd(ctx context.Context, key string, els ...interface{}) *redis.IntCmd {
	cmd := Client.PFAdd(ctx, key, els...)
	logError(cmd)
	return cmd
}
func PFCount(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.PFCount(ctx, keys...)
	logError(cmd)
	return cmd
}
func PFMerge(ctx context.Context, dest string, keys ...string) *redis.StatusCmd {
	cmd := Client.PFMerge(ctx, dest, keys...)
	logError(cmd)
	return cmd
}

func BgRewriteAOF(ctx context.Context) *redis.StatusCmd {
	cmd := Client.BgRewriteAOF(ctx)
	logError(cmd)
	return cmd
}
func BgSave(ctx context.Context) *redis.StatusCmd {
	cmd := Client.BgSave(ctx)
	logError(cmd)
	return cmd
}
func ClientKill(ctx context.Context, ipPort string) *redis.StatusCmd {
	cmd := Client.ClientKill(ctx, ipPort)
	logError(cmd)
	return cmd
}
func ClientKillByFilter(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := Client.ClientKillByFilter(ctx, keys...)
	logError(cmd)
	return cmd
}
func ClientList(ctx context.Context) *redis.StringCmd {
	cmd := Client.ClientList(ctx)
	logError(cmd)
	return cmd
}
func ClientPause(ctx context.Context, dur time.Duration) *redis.BoolCmd {
	cmd := Client.ClientPause(ctx, dur)
	logError(cmd)
	return cmd
}
func ClientID(ctx context.Context) *redis.IntCmd {
	cmd := Client.ClientID(ctx)
	logError(cmd)
	return cmd
}
func ConfigGet(ctx context.Context, parameter string) *redis.SliceCmd {
	cmd := Client.ConfigGet(ctx, parameter)
	logError(cmd)
	return cmd
}
func ConfigResetStat(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ConfigResetStat(ctx)
	logError(cmd)
	return cmd
}
func ConfigSet(ctx context.Context, parameter, value string) *redis.StatusCmd {
	cmd := Client.ConfigSet(ctx, parameter, value)
	logError(cmd)
	return cmd
}
func ConfigRewrite(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ConfigRewrite(ctx)
	logError(cmd)
	return cmd
}
func DBSize(ctx context.Context) *redis.IntCmd {
	cmd := Client.DBSize(ctx)
	logError(cmd)
	return cmd
}
func FlushAll(ctx context.Context) *redis.StatusCmd {
	cmd := Client.FlushAll(ctx)
	logError(cmd)
	return cmd
}
func FlushAllAsync(ctx context.Context) *redis.StatusCmd {
	cmd := Client.FlushAllAsync(ctx)
	logError(cmd)
	return cmd
}
func FlushDB(ctx context.Context) *redis.StatusCmd {
	cmd := Client.FlushDB(ctx)
	logError(cmd)
	return cmd
}
func FlushDBAsync(ctx context.Context) *redis.StatusCmd {
	cmd := Client.FlushDBAsync(ctx)
	logError(cmd)
	return cmd
}
func Info(ctx context.Context, section ...string) *redis.StringCmd {
	cmd := Client.Info(ctx, section...)
	logError(cmd)
	return cmd
}
func LastSave(ctx context.Context) *redis.IntCmd {
	cmd := Client.LastSave(ctx)
	logError(cmd)
	return cmd
}
func Save(ctx context.Context) *redis.StatusCmd {
	cmd := Client.Save(ctx)
	logError(cmd)
	return cmd
}
func Shutdown(ctx context.Context) *redis.StatusCmd {
	cmd := Client.Shutdown(ctx)
	logError(cmd)
	return cmd
}
func ShutdownSave(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ShutdownSave(ctx)
	logError(cmd)
	return cmd
}
func ShutdownNoSave(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ShutdownNoSave(ctx)
	logError(cmd)
	return cmd
}
func SlaveOf(ctx context.Context, host, port string) *redis.StatusCmd {
	cmd := Client.SlaveOf(ctx, host, port)
	logError(cmd)
	return cmd
}
func Time(ctx context.Context) *redis.TimeCmd {
	cmd := Client.Time(ctx)
	logError(cmd)
	return cmd
}
func DebugObject(ctx context.Context, key string) *redis.StringCmd {
	cmd := Client.DebugObject(ctx, key)
	logError(cmd)
	return cmd
}
func ReadOnly(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ReadOnly(ctx)
	logError(cmd)
	return cmd
}
func ReadWrite(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ReadWrite(ctx)
	logError(cmd)
	return cmd
}
func MemoryUsage(ctx context.Context, key string, samples ...int) *redis.IntCmd {
	cmd := Client.MemoryUsage(ctx, key, samples...)
	logError(cmd)
	return cmd
}

func Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	cmd := Client.Eval(ctx, script, keys, args...)
	logError(cmd)
	return cmd
}
func EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	cmd := Client.EvalSha(ctx, sha1, keys, args...)
	logError(cmd)
	return cmd
}
func ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd {
	cmd := Client.ScriptExists(ctx, hashes...)
	logError(cmd)
	return cmd
}
func ScriptFlush(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ScriptFlush(ctx)
	logError(cmd)
	return cmd
}
func ScriptKill(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ScriptKill(ctx)
	logError(cmd)
	return cmd
}
func ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	cmd := Client.ScriptLoad(ctx, script)
	logError(cmd)
	return cmd
}

func Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	cmd := Client.Publish(ctx, channel, message)
	logError(cmd)
	return cmd
}
func PubSubChannels(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := Client.PubSubChannels(ctx, pattern)
	logError(cmd)
	return cmd
}
func PubSubNumSub(ctx context.Context, channels ...string) *redis.StringIntMapCmd {
	cmd := Client.PubSubNumSub(ctx, channels...)
	logError(cmd)
	return cmd
}
func PubSubNumPat(ctx context.Context) *redis.IntCmd {
	cmd := Client.PubSubNumPat(ctx)
	logError(cmd)
	return cmd
}

func ClusterSlots(ctx context.Context) *redis.ClusterSlotsCmd {
	cmd := Client.ClusterSlots(ctx)
	logError(cmd)
	return cmd
}
func ClusterNodes(ctx context.Context) *redis.StringCmd {
	cmd := Client.ClusterNodes(ctx)
	logError(cmd)
	return cmd
}
func ClusterMeet(ctx context.Context, host, port string) *redis.StatusCmd {
	cmd := Client.ClusterMeet(ctx, host, port)
	logError(cmd)
	return cmd
}
func ClusterForget(ctx context.Context, nodeID string) *redis.StatusCmd {
	cmd := Client.ClusterForget(ctx, nodeID)
	logError(cmd)
	return cmd
}
func ClusterReplicate(ctx context.Context, nodeID string) *redis.StatusCmd {
	cmd := Client.ClusterReplicate(ctx, nodeID)
	logError(cmd)
	return cmd
}
func ClusterResetSoft(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ClusterResetSoft(ctx)
	logError(cmd)
	return cmd
}
func ClusterResetHard(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ClusterResetHard(ctx)
	logError(cmd)
	return cmd
}
func ClusterInfo(ctx context.Context) *redis.StringCmd {
	cmd := Client.ClusterInfo(ctx)
	logError(cmd)
	return cmd
}
func ClusterKeySlot(ctx context.Context, key string) *redis.IntCmd {
	cmd := Client.ClusterKeySlot(ctx, key)
	logError(cmd)
	return cmd
}
func ClusterGetKeysInSlot(ctx context.Context, slot int, count int) *redis.StringSliceCmd {
	cmd := Client.ClusterGetKeysInSlot(ctx, slot, count)
	logError(cmd)
	return cmd
}
func ClusterCountFailureReports(ctx context.Context, nodeID string) *redis.IntCmd {
	cmd := Client.ClusterCountFailureReports(ctx, nodeID)
	logError(cmd)
	return cmd
}
func ClusterCountKeysInSlot(ctx context.Context, slot int) *redis.IntCmd {
	cmd := Client.ClusterCountKeysInSlot(ctx, slot)
	logError(cmd)
	return cmd
}
func ClusterDelSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	cmd := Client.ClusterDelSlots(ctx, slots...)
	logError(cmd)
	return cmd
}
func ClusterDelSlotsRange(ctx context.Context, min, max int) *redis.StatusCmd {
	cmd := Client.ClusterDelSlotsRange(ctx, min, max)
	logError(cmd)
	return cmd
}
func ClusterSaveConfig(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ClusterSaveConfig(ctx)
	logError(cmd)
	return cmd
}
func ClusterSlaves(ctx context.Context, nodeID string) *redis.StringSliceCmd {
	cmd := Client.ClusterSlaves(ctx, nodeID)
	logError(cmd)
	return cmd
}
func ClusterFailover(ctx context.Context) *redis.StatusCmd {
	cmd := Client.ClusterFailover(ctx)
	logError(cmd)
	return cmd
}
func ClusterAddSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	cmd := Client.ClusterAddSlots(ctx, slots...)
	logError(cmd)
	return cmd
}
func ClusterAddSlotsRange(ctx context.Context, min, max int) *redis.StatusCmd {
	cmd := Client.ClusterAddSlotsRange(ctx, min, max)
	logError(cmd)
	return cmd
}

func GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	cmd := Client.GeoAdd(ctx, key, geoLocation...)
	logError(cmd)
	return cmd
}
func GeoPos(ctx context.Context, key string, members ...string) *redis.GeoPosCmd {
	cmd := Client.GeoPos(ctx, key, members...)
	logError(cmd)
	return cmd
}
func GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	cmd := Client.GeoRadius(ctx, key, longitude, latitude, query)
	logError(cmd)
	return cmd
}
func GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.IntCmd {
	cmd := Client.GeoRadiusStore(ctx, key, longitude, latitude, query)
	logError(cmd)
	return cmd
}
func GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	cmd := Client.GeoRadiusByMember(ctx, key, member, query)
	logError(cmd)
	return cmd
}
func GeoRadiusByMemberStore(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) *redis.IntCmd {
	cmd := Client.GeoRadiusByMemberStore(ctx, key, member, query)
	logError(cmd)
	return cmd
}
func GeoDist(ctx context.Context, key string, member1, member2, unit string) *redis.FloatCmd {
	cmd := Client.GeoDist(ctx, key, member1, member2, unit)
	logError(cmd)
	return cmd
}
func GeoHash(ctx context.Context, key string, members ...string) *redis.StringSliceCmd {
	cmd := Client.GeoHash(ctx, key, members...)
	logError(cmd)
	return cmd
}

func logError(cmd redis.Cmder) {
	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		log.WithField("cmd", cmd.Args()).WithError(cmd.Err()).Warn(cmd.String())
	}
}

func SimpleLock(lockid string) error {
	i := 0
	for {
		vok, err := Client.SetNX(context.Background(), lockid, time.Now().Unix(), time.Second*30).Result()
		if vok {
			return nil
		}
		DxCommonLib.Sleep(time.Millisecond * 100)
		if i == 99 {
			if err == nil {
				err = ErrLockerTimeout
			}
			return err
		}
		i++
	}
}

func SimpleUnLock(lockid string) {
	val, _ := Client.Del(context.Background(), lockid).Result()
	if val != 1 {
		DxCommonLib.MustRunAsync(dosimpleUnLock, lockid)
	}
}

func dosimpleUnLock(data ...interface{}) {
	lockid := data[0].(string)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		val, err := Client.Del(ctx, lockid).Result()
		if val == 1 {
			return
		}
		if i == 9 && err != nil {
			log.WithError(err).Errorf("redis SimpleUnLock %s 错误", lockid)
			return
		}
		DxCommonLib.Sleep(time.Millisecond * 200)
	}
}

type Lock struct {
	lockKey string
}

func TryLock(key string, timeOut time.Duration) *Lock {
	erridx := 0
	timeout := DxCommonLib.After(timeOut)
	for {
		select {
		case <-timeout:
			return nil
		default:
			vbool := Client.SetNX(context.Background(), key, time.Now().Unix(), time.Second*30)
			if ok, err := vbool.Result(); err == nil {
				if ok {
					return &Lock{key}
				}
				//失败
			} else {
				log.WithError(err).Error("redis锁发生错误")
				erridx++
				if erridx > 10 {
					return nil
				}
			}
			DxCommonLib.Sleep(time.Millisecond * 500)
		}
	}
}

func UnLock(lock *Lock) (bool, error) {
	if lock == nil {
		return false, nil
	}
	tk := DxCommonLib.After(time.Second * 15)
	for {
		select {
		case <-tk:
			return false, ErrLockerTimeout
		default:
			val, err := Client.Del(context.Background(), lock.lockKey).Result()
			if err != nil {
				log.WithError(err).Error("redis unlock 错误")
			} else if val == 1 {
				return true, nil
			}
		}
	}
}
