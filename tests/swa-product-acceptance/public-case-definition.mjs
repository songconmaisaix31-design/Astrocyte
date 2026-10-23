export const publicSource = 'https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats.git'
export const publicRevision = 'da3bdddb085808863090a05f60cc1c64415039cf'
export const publicFiles = ['README.md', 'go.mod', 'stats.go', 'stats_test.go']

// 准备和发送前校验共用同一份固定公开正文。
export function publicCase(workers, originals) {
  const context = JSON.stringify({ source: publicSource, revision: publicRevision, originals })
  const roles = [
    '核对四份原文的模块组织、现有 Mean API 和真实测试覆盖，按文件引用已有事实与缺少的证据。',
    '审阅 Median 的设计条件，覆盖奇偶长度、负数、重复值、输入顺序保持与现有 Mean 兼容，不把建议写成已有实现。',
    '审阅空输入、NaN、正负 Infinity、有限大中间值溢出等失败与边界条件，区分现有事实、建议和需要实际运行的检查。',
    '独立核对现有测试能证明什么、还缺什么；分别指出代码事实、未执行检查和建议，收到后续输入时在原会话复核。',
    '核对复杂度、是否修改输入和实现选择的条件，按原文说明依据；无法由四份原文证实的行为保持未知。',
    '独立核对任务成功标准和交付可审性，检查事实、设计建议、失败边界与尚未运行证据是否分别表述。',
  ]
  return {
    inputs: Array.from({ length: workers }, (_, index) => {
      const name = `worker-${index + 1}`
      const instruction = `你负责本次公开来源审阅的第${index + 1}项。${roles[index % roles.length]}逐条给出推理与来源位置，检查其他可能解释；收到后续输入时在同一会话继续。禁止调用工具、读写文件或访问其他资料；不要声称运行了测试。部分 Worker 会被计划停止，保留真实停止状态。\n`
      return { name, file: `${name}.txt`, body: instruction + context + '\n' }
    }),
    spec: {
      goal: '审阅固定公开 stats 项目，取得多个真实 Worker 的来源核对、设计与失败路径分析；不修改程序。',
      source_repo: publicSource, source_revision: publicRevision, max_workers: workers,
      success_criteria: [
        `${workers}个实际产品 Worker 有独立 Dispatch/session 且真实模型执行区间重叠；不以进程存在或排队代替。`,
        `指定停止一项时其仍真实运行，其余至少${workers - 1}项保留原身份继续；一项在同会话处理后续输入。`,
        '现有事实与设计建议分别表述并引用四份固定原文；未知、未运行测试和被停止项保持原状态。',
        '取得实际完整来源成果，确认全部所属资源退出并保护非所属进程、原账号和数据。',
      ],
      context: { references: [], files: publicFiles },
    },
    followup: '请在原会话继续，复核你的首次回答中各项来源与边界条件，纠正无法由固定原文证明的陈述；结合本会话已有真实消息形成完整说明。禁止调用工具，不读取其他资料，不声称执行了测试。\n' + context + '\n',
  }
}
