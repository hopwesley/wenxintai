/* === 题库 === */
interface Item {
    id: number;
    type: 'D' | 'A' | 'V';
    rev: boolean;
    text: string;
}

const ITEMS: Item[] = [
    { id: 1, type: 'D', rev: false, text: '最近两周，我常感到情绪低落、提不起兴趣完成课程或社团事务。' },
    { id: 2, type: 'A', rev: false, text: '接近考试/作业DDL时，我会出现紧张、心慌或坐立不安的感觉。' },
    { id: 3, type: 'D', rev: false, text: '我常对未来学业或就业前景感到悲观，无力规划。' },
    { id: 4, type: 'A', rev: false, text: '在人多或陌生场合发言时，我会明显担忧他人评价。' },
    { id: 5, type: 'D', rev: true, text: '完成日常学习任务后，我仍能感到开心并从中获得满足。' },
    { id: 6, type: 'A', rev: false, text: '我容易因小事紧张，身体出现肌肉僵硬、出汗、心跳加快等反应。' },
    { id: 7, type: 'D', rev: false, text: '近两周，我的睡眠质量下降（入睡困难、早醒或睡不解乏）。' },
    { id: 8, type: 'A', rev: true, text: '面对不确定安排（课程调整/突发任务），我能保持放松与从容。' },
    { id: 9, type: 'D', rev: false, text: '我常陷入自我否定或反复回想负面经历而难以抽离。' },
    { id: 10, type: 'A', rev: false, text: '我经常为无法完全掌控学习细节而焦虑或烦躁。' },
    { id: 11, type: 'D', rev: false, text: '食欲与能量较以往下降，做事需要很大努力才能坚持。' },
    { id: 12, type: 'A', rev: false, text: '我常担心与同学关系或人际误解，并因此紧张不安。' },
    { id: 13, type: 'V', rev: false, text: '我从不感到任何压力或担忧（完全没有过）。' },
    { id: 14, type: 'V', rev: false, text: '我总是心情愉快，从不出现消极想法。' },
    { id: 15, type: 'V', rev: true, text: '我偶尔也会状态不好或做事拖延（这很正常）。' },
    { id: 16, type: 'V', rev: false, text: '我经常在1分钟内就能完全搞定所有课程任务。' }
];

const qWrap = document.getElementById('questions') as HTMLElement;

function renderQuestions(): void {
    qWrap.innerHTML = '';
    ITEMS.forEach(item => {
        const div = document.createElement('div');
        div.className = 'q';
        const hint = item.type === 'V' ? '（效度题）' : '';
        div.innerHTML = `
      <div class="qtitle">${item.id}. ${item.text} <span class="muted">${hint}</span></div>
      <div class="scale" role="radiogroup" aria-label="题目${item.id}">
        ${[1,2,3,4,5].map(v=>`
          <label><input type="radio" name="q${item.id}" value="${v}"><span>${v}</span></label>`).join('')}
      </div>`;
        qWrap.appendChild(div);
    });
}
renderQuestions();

const clamp = (v: number, min: number, max: number): number => Math.max(min, Math.min(max, v));
const reverse5 = (v: number): number => ({ 1: 5, 2: 4, 3: 3, 4: 2, 5: 1 }[v] || v);

function readAnswers(): Record<number, number | null> {
    const ans: Record<number, number | null> = {};
    for (const it of ITEMS) {
        const el = document.querySelector<HTMLInputElement>(`input[name=q${it.id}]:checked`);
        ans[it.id] = el ? +el.value : null;
    }
    return ans;
}

function scoreAndJudge(ans: Record<number, number | null>) {
    const missing = Object.values(ans).filter(v => v == null).length;
    let imputed = { ...ans };

    if (missing > 2) return { error: '作答缺失较多（>2题），建议重新作答。' };

    for (const k in imputed) {
        if (imputed[k as any] == null) imputed[k as any] = 3;
    }

    const calc = (type: 'D' | 'A') => {
        const raw = ITEMS.filter(it => it.type === type)
            .map(it => it.rev ? reverse5(imputed[it.id]!) : imputed[it.id]!);
        const sum = raw.reduce((a, b) => a + b, 0), n = raw.length;
        const pct = Math.round(((sum - n * 1) / (n * 4)) * 100);
        return { sum, n, pct: clamp(pct, 0, 100) };
    };

    const D = calc('D'), A = calc('A');

    const v13 = imputed[13], v14 = imputed[14], v16 = imputed[16], v15 = imputed[15];
    let vFlags: string[] = [];
    if (v13! >= 4) vFlags.push('不常见肯定（13）');
    if (v14! >= 4) vFlags.push('过度正面（14）');
    if (v16! >= 4) vFlags.push('不合常理速度（16）');
    if (v15! <= 2) vFlags.push('否认普遍现象（15）');

    const validityRisk = vFlags.length >= 2 ? '效度可疑'
        : (vFlags.length === 1 ? '效度轻微可疑' : '效度可接受');

    function band(p: number) {
        if (p < 40) return { lvl: '低', cls: 'badge-ok' };
        if (p < 55) return { lvl: '轻度', cls: 'badge-warn' };
        if (p < 70) return { lvl: '中度', cls: 'badge-warn' };
        return { lvl: '较高', cls: 'badge-bad' };
    }

    const bd = band(D.pct), ba = band(A.pct);

    const studentType = (document.querySelector<HTMLInputElement>('input[name=studentType]:checked') || {}).value || '未说明';
    const gender = (document.querySelector<HTMLInputElement>('input[name=gender]:checked') || {}).value || '未说明';

    const words = buildReport({ studentType, gender, D, A, validityRisk, vFlags });
    return { D, A, bd, ba, validityRisk, vFlags, words };
}

function buildReport(ctx: {
    studentType: string;
    gender: string;
    D: { pct: number };
    A: { pct: number };
    validityRisk: string;
    vFlags: string[];
}): string {
    const { studentType, gender, D, A, validityRisk, vFlags } = ctx;
    const riskFocus = (D.pct >= 70 || A.pct >= 70)
        ? '存在较高风险'
        : ((D.pct >= 55 || A.pct >= 55) ? '存在一定风险' : '整体风险较低');
    const lines: string[] = [];
    lines.push(`【对象概况】受测者（类别：${studentType}，性别：${gender}）完成抑郁/焦虑筛查量表，共16题（含效度题）。作答完整度良好。`);
    lines.push(`【核心指标】抑郁倾向风险评分 ${D.pct} / 100，焦虑倾向风险评分 ${A.pct} / 100。总体判断：${riskFocus}。`);
    if (validityRisk !== '效度可接受') {
        lines.push(`【作答效度】当前报告显示“${validityRisk}”（提示项：${vFlags.join('、')}），建议结合访谈或复测核实。`);
    } else {
        lines.push(`【作答效度】效度指标在可接受范围内，本次结果可用于初步筛查参考。`);
    }
    lines.push(`【使用范围】本报告仅用于校方风险筛查与资源分配，不构成临床诊断。`);
    return lines.join('\n');
}

function setStep(n: number): void {
    document.querySelectorAll('.steps .step').forEach((el, i) => {
        el.classList.toggle('active', i === n - 1);
    });
}

(document.getElementById('toStep1') as HTMLButtonElement).onclick = () => {
    (document.getElementById('step0') as HTMLElement).hidden = true;
    (document.getElementById('step1') as HTMLElement).hidden = false;
    setStep(2);
    window.scrollTo({ top: 0, behavior: 'smooth' });
};

(document.getElementById('submit') as HTMLButtonElement).onclick = () => {
    const ans = readAnswers(), res: any = scoreAndJudge(ans);
    if (res.error) { alert(res.error); return; }

    const kpi = document.getElementById('kpi') as HTMLElement;
    kpi.innerHTML = `
      <div class="item item-d"><div class="muted">抑郁风险（0~100）</div><div class="score">${res.D.pct}</div><div class="${res.bd.cls}">分层：${res.bd.lvl}</div></div>
      <div class="item item-a"><div class="muted">焦虑风险（0~100）</div><div class="score">${res.A.pct}</div><div class="${res.ba.cls}">分层：${res.ba.lvl}</div></div>
      <div class="item item-v"><div class="muted">作答效度</div><div class="score">${res.validityRisk}</div><div class="muted">${res.vFlags.length?('提示项：'+res.vFlags.join('、')):'无异常提示'}</div></div>`;

    (document.getElementById('report') as HTMLTextAreaElement).value = res.words;
    (document.getElementById('validity') as HTMLElement).innerHTML = `<p class="muted">判定逻辑：反向计分（5,8,15题），效度题≥2项提示则效度可疑。</p>`;

    (document.getElementById('step1') as HTMLElement).hidden = true;
    (document.getElementById('step2') as HTMLElement).hidden = false;
    setStep(3);
    window.scrollTo({ top: 0, behavior: 'smooth' });
};

(document.getElementById('back') as HTMLButtonElement).onclick = () => {
    (document.getElementById('step2') as HTMLElement).hidden = true;
    (document.getElementById('step1') as HTMLElement).hidden = false;
    setStep(2);
    window.scrollTo({ top: 0, behavior: 'smooth' });
};

(document.getElementById('print') as HTMLButtonElement).onclick = () => window.print();
