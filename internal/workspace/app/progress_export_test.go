package app

// 外部测试调用同一生产函数，并使用实际 SQLite 生命周期触发错误。
var SettleProgressFailureForTest = (*projectProgressService).failSettle
var UnknownComputedProgressForTest = (*projectProgressService).unknownComputed
