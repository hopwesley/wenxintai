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
