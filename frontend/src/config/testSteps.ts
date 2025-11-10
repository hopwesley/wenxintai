export const STEPS = {
  basic: [
    { key: 'basicInfo', titleKey: 'steps.basic.info' },
    { key: 'stage1Questions', titleKey: 'steps.basic.stage1' },
    { key: 'stage2Questions', titleKey: 'steps.basic.stage2' },
    { key: 'report', titleKey: 'steps.basic.report' },
  ],
  pro: [
    { key: 'basicInfo', titleKey: 'steps.pro.info' },
    { key: 'stage1A', titleKey: 'steps.pro.stage1a' },
    { key: 'stage1B', titleKey: 'steps.pro.stage1b' },
    { key: 'stage2A', titleKey: 'steps.pro.stage2a' },
    { key: 'stage2B', titleKey: 'steps.pro.stage2b' },
    { key: 'report', titleKey: 'steps.pro.report' },
  ],
} as const

export type Variant = keyof typeof STEPS

export function isVariant(input: string): input is Variant {
  return Object.prototype.hasOwnProperty.call(STEPS, input)
}
