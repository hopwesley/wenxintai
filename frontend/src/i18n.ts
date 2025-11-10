import { App, Ref, inject, ref } from 'vue'

const I18N_SYMBOL = Symbol('i18n')

type MessageTree = Record<string, string | MessageTree>

type Messages = Record<string, MessageTree>

export interface I18nContext {
  locale: Ref<string>
  t: (key: string) => string
}

const defaultMessages: Messages = {
  zh: {
    steps: {
      basic: {
        info: '填写基础资料',
        stage1: '第一阶段测试题目',
        stage2: '第二阶段测试题目',
        report: '完成测试',
      },
      pro: {
        info: '填写基础资料',
        stage1a: '第一阶段A',
        stage1b: '第一阶段B',
        stage2a: '第二阶段A',
        stage2b: '第二阶段B',
        report: '完成测试',
      },
    },
    form: {
      age: {
        label: '出生年龄',
        placeholder: '选择您的年龄',
      },
      mode: {
        label: '选科模式',
        placeholder: '选择所学科目',
      },
      hobby: {
        label: '兴趣爱好',
        placeholder: '选择您的兴趣爱好',
      },
      validation: {
        required: '请填写此字段',
        ageRange: '请输入合理的年龄',
      },
    },
    btn: {
      start: '开始',
      prev: '返回上一页',
      next: '下一页',
      submit: '提交',
    },
    scale: {
      never: '从不',
      rare: '很少',
      normal: '一般',
      often: '经常',
      alot: '非常多',
    },
    loading: {
      default: '加载中…',
      submitting: '提交中…',
    },
    error: {
      network: '网络异常，请重试',
      noQuestions: '暂无题目，请稍后再试',
    },
    placeholder: {
      title: '功能开发中',
      description: '专业版该阶段尚未开放，敬请期待。',
      back: '返回首页',
    },
    report: {
      title: 'AI 报告',
      loading: '正在生成报告…',
      empty: '暂无报告内容',
    },
    disclaimer: '免责声明 基于AI生成，仅供参考',
    invite: {
      freeHint: '使用邀请码免费体验完整测评',
      title: '请输入邀请码',
      description: '每个邀请码仅可使用一次，请确认后提交。',
      placeholder: '邀请码',
      confirm: '确认开始',
      cancel: '取消',
      loading: '验证中…',
      errors: {
        empty: '请输入邀请码',
        used: '该邀请码已被使用',
        expired: '该邀请码已过期',
        not_found: '邀请码无效',
        network: '网络异常，请重试',
        unknown: '发生未知错误，请稍后再试'
      }
    }
  },
  en: {
    steps: {
      basic: {
        info: 'Fill in basic information',
        stage1: 'Stage 1 questions',
        stage2: 'Stage 2 questions',
        report: 'View report',
      },
      pro: {
        info: 'Fill in basic information',
        stage1a: 'Stage 1 · A',
        stage1b: 'Stage 1 · B',
        stage2a: 'Stage 2 · A',
        stage2b: 'Stage 2 · B',
        report: 'View report',
      },
    },
    form: {
      age: {
        label: 'Age',
        placeholder: 'Select your age',
      },
      mode: {
        label: 'Subject scheme',
        placeholder: 'Choose the scheme you follow',
      },
      hobby: {
        label: 'Hobby',
        placeholder: 'Select your hobby',
      },
      validation: {
        required: 'This field is required',
        ageRange: 'Please enter a valid age',
      },
    },
    btn: {
      start: 'Start',
      prev: 'Previous',
      next: 'Next',
      submit: 'Submit',
    },
    scale: {
      never: 'Never',
      rare: 'Rarely',
      normal: 'Sometimes',
      often: 'Often',
      alot: 'Very often',
    },
    loading: {
      default: 'Loading…',
      submitting: 'Submitting…',
    },
    error: {
      network: 'Network error, please try again.',
      noQuestions: 'No questions available for now.',
    },
    placeholder: {
      title: 'Coming soon',
      description: 'This professional stage is under construction. Stay tuned!',
      back: 'Back to home',
    },
    report: {
      title: 'AI Report',
      loading: 'Generating report…',
      empty: 'Report is not available yet.',
    },
    disclaimer: 'Disclaimer: AI-generated content for reference only.',
    invite: {
      freeHint: 'Start the full assessment for free with an invite code.',
      title: 'Enter Invite Code',
      description: 'Each invite code can be used only once. Submit carefully.',
      placeholder: 'Invite code',
      confirm: 'Start now',
      cancel: 'Cancel',
      loading: 'Verifying…',
      errors: {
        empty: 'Please enter the invite code.',
        used: 'This invite code has already been used.',
        expired: 'This invite code has expired.',
        not_found: 'Invalid invite code.',
        network: 'Network error, please try again.',
        unknown: 'Unexpected error, please try later.'
      }
    }
  }
}

export interface CreateI18nOptions {
  locale?: string
  messages?: Messages
}

export function createI18n(options: CreateI18nOptions = {}) {
  const locale = ref(options.locale ?? 'zh')
  const messages = options.messages ?? defaultMessages

  const ctx: I18nContext = {
    locale,
    t(key: string) {
      const targetLocale = messages[locale.value] ?? {}
      const segments = key.split('.')
      let current: string | MessageTree | undefined = targetLocale

      for (const segment of segments) {
        if (typeof current !== 'object' || current === null) {
          return key
        }
        current = current[segment]
      }

      return typeof current === 'string' ? current : key
    }
  }

  function install(app: App) {
    app.provide(I18N_SYMBOL, ctx)
  }

  return { install, ...ctx }
}

export function useI18n(): I18nContext {
  const ctx = inject<I18nContext>(I18N_SYMBOL)
  if (!ctx) {
    throw new Error('i18n context is not available. Make sure createI18n() is installed.')
  }
  return ctx
}
