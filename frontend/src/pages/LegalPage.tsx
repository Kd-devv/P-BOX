import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { ArrowLeft, Shield, ScrollText, Lock, Scale, AlertTriangle } from 'lucide-react'

export default function LegalPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const { themeStyle } = useThemeStore()
  const isZh = i18n.language.startsWith('zh')

  return (
    <div className="space-y-6 pb-10">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          className={cn(
            'p-2 rounded-lg transition-colors',
            themeStyle === 'apple-glass'
              ? 'hover:bg-black/5 text-slate-600'
              : 'hover:bg-white/10 text-slate-400'
          )}
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className={cn(
            'text-2xl font-bold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {t('legal.title')}
          </h1>
          <p className="text-sm text-slate-500">
            {isZh ? '法律声明与服务条款' : 'Legal Disclaimer & Terms of Service'}
          </p>
        </div>
      </div>

      {/* Important Notice */}
      <div className={cn(
        'glass-card p-5 border-l-4 border-amber-500',
        themeStyle === 'apple-glass' ? 'bg-amber-50/50' : 'bg-amber-500/10'
      )}>
        <div className="flex items-start gap-3">
          <AlertTriangle className="w-6 h-6 text-amber-500 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className={cn(
              'font-semibold mb-2',
              themeStyle === 'apple-glass' ? 'text-amber-700' : 'text-amber-400'
            )}>
              {isZh ? '重要提示' : 'Important Notice'}
            </h3>
            <p className={cn(
              'text-sm leading-relaxed',
              themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
            )}>
              {isZh 
                ? '欢迎使用 P-BOX。在使用本软件之前，请您仔细阅读并充分理解以下条款。您访问、下载、安装或使用本软件的行为，即表示您已阅读、理解并同意接受以下所有条款和条件的约束。如果您不同意这些条款，请立即停止使用本软件并将其从您的设备中删除。'
                : 'Welcome to P-BOX. Before using this software, please read and fully understand the following terms. By accessing, downloading, installing, or using this software, you acknowledge that you have read, understood, and agree to be bound by all the terms and conditions below. If you do not agree to these terms, please immediately stop using this software and remove it from your device.'
              }
            </p>
          </div>
        </div>
      </div>

      {/* Section 1: Software Nature */}
      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="app-icon blue">
            <ScrollText className="w-4 h-4" />
          </div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {isZh ? '第一条 软件性质与授权' : 'Article 1: Software Nature & License'}
          </h2>
        </div>
        <div className={cn(
          'space-y-3 text-sm leading-relaxed',
          themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
        )}>
          {isZh ? (
            <>
              <p><strong>1.1</strong> 本软件是一款开源、免费的网络代理管理工具，基于开源协议发布，旨在为用户提供网络流量管理、代理配置和网络优化等功能。</p>
              <p><strong>1.2</strong> 本软件的源代码公开可用，任何个人或组织均可在遵守开源协议的前提下，自由使用、复制、修改和分发本软件。</p>
              <p><strong>1.3</strong> 本软件完全免费，任何以本软件名义进行的收费行为均与本软件作者和开发团队无关。如您通过付费方式获得本软件，请向相关销售方追究责任。</p>
              <p><strong>1.4</strong> 本软件不提供任何形式的网络代理服务、VPN服务或翻墙服务，仅作为本地代理管理工具使用。</p>
            </>
          ) : (
            <>
              <p><strong>1.1</strong> This software is an open-source, free network proxy management tool released under an open-source license, designed to provide users with network traffic management, proxy configuration, and network optimization features.</p>
              <p><strong>1.2</strong> The source code of this software is publicly available. Any individual or organization may freely use, copy, modify, and distribute this software in compliance with the open-source license.</p>
              <p><strong>1.3</strong> This software is completely free. Any charging behavior in the name of this software has nothing to do with the author and development team. If you obtained this software through payment, please hold the seller accountable.</p>
              <p><strong>1.4</strong> This software does not provide any form of network proxy service, VPN service, or circumvention service, and is only used as a local proxy management tool.</p>
            </>
          )}
        </div>
      </div>

      {/* Section 2: Usage Restrictions */}
      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="app-icon red">
            <Scale className="w-4 h-4" />
          </div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {isZh ? '第二条 使用限制与合规要求' : 'Article 2: Usage Restrictions & Compliance'}
          </h2>
        </div>
        <div className={cn(
          'space-y-4 text-sm leading-relaxed',
          themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
        )}>
          {isZh ? (
            <>
              <div>
                <p className="font-medium mb-2">2.1 地区法律合规</p>
                <p>用户在使用本软件时，必须严格遵守其所在国家或地区的所有适用法律、法规、规章及其他具有法律效力的规范性文件。用户有责任自行了解并确保其使用行为符合当地法律要求。</p>
              </div>
              <div>
                <p className="font-medium mb-2">2.2 禁止行为</p>
                <p className="mb-2">用户在使用本软件时，不得从事以下任何行为：</p>
                <ul className="list-disc list-inside space-y-1 ml-2">
                  <li>违反国家法律法规、危害国家安全、泄露国家秘密的行为</li>
                  <li>侵犯他人知识产权、商业秘密或其他合法权益的行为</li>
                  <li>传播淫秽、色情、赌博、暴力、恐怖或教唆犯罪的信息</li>
                  <li>散布谣言、扰乱社会秩序、破坏社会稳定的行为</li>
                  <li>从事任何形式的网络攻击、入侵、破坏活动</li>
                  <li>未经授权访问、使用他人计算机系统或网络</li>
                  <li>规避、破坏任何安全措施或访问控制机制</li>
                  <li>侵犯他人隐私、窃取他人个人信息</li>
                  <li>从事任何形式的欺诈、诈骗活动</li>
                  <li>其他任何违反法律法规或公序良俗的行为</li>
                </ul>
              </div>
              <div>
                <p className="font-medium mb-2">2.3 使用目的</p>
                <p className="mb-2">本软件仅供以下合法目的使用：</p>
                <ul className="list-disc list-inside space-y-1 ml-2">
                  <li>网络技术学习与研究</li>
                  <li>软件开发与测试</li>
                  <li>企业内部网络管理</li>
                  <li>合法的跨境业务需求</li>
                  <li>其他符合法律规定的正当用途</li>
                </ul>
              </div>
            </>
          ) : (
            <>
              <div>
                <p className="font-medium mb-2">2.1 Regional Legal Compliance</p>
                <p>When using this software, users must strictly comply with all applicable laws, regulations, rules, and other legally binding normative documents in their country or region. Users are responsible for understanding and ensuring that their use complies with local legal requirements.</p>
              </div>
              <div>
                <p className="font-medium mb-2">2.2 Prohibited Activities</p>
                <p className="mb-2">Users shall not engage in any of the following activities when using this software:</p>
                <ul className="list-disc list-inside space-y-1 ml-2">
                  <li>Violating national laws, endangering national security, or leaking state secrets</li>
                  <li>Infringing on others' intellectual property, trade secrets, or other legitimate rights</li>
                  <li>Spreading obscene, pornographic, gambling, violent, terrorist, or crime-instigating information</li>
                  <li>Spreading rumors, disrupting social order, or undermining social stability</li>
                  <li>Engaging in any form of cyber attack, intrusion, or sabotage</li>
                  <li>Unauthorized access to or use of others' computer systems or networks</li>
                  <li>Circumventing or destroying any security measures or access control mechanisms</li>
                  <li>Violating others' privacy or stealing personal information</li>
                  <li>Engaging in any form of fraud or scam</li>
                  <li>Any other activities that violate laws, regulations, or public order and good morals</li>
                </ul>
              </div>
              <div>
                <p className="font-medium mb-2">2.3 Permitted Uses</p>
                <p className="mb-2">This software is only for the following legitimate purposes:</p>
                <ul className="list-disc list-inside space-y-1 ml-2">
                  <li>Network technology learning and research</li>
                  <li>Software development and testing</li>
                  <li>Internal enterprise network management</li>
                  <li>Legitimate cross-border business needs</li>
                  <li>Other legitimate purposes in compliance with laws</li>
                </ul>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Section 3: Disclaimer */}
      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="app-icon orange">
            <Shield className="w-4 h-4" />
          </div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {isZh ? '第三条 免责声明' : 'Article 3: Disclaimer'}
          </h2>
        </div>
        <div className={cn(
          'space-y-3 text-sm leading-relaxed',
          themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
        )}>
          {isZh ? (
            <>
              <p><strong>3.1 "按现状"提供</strong><br/>本软件按"现状"（AS IS）和"可用"（AS AVAILABLE）基础提供，不附带任何形式的明示或暗示保证，包括但不限于对适销性、特定用途适用性、非侵权性、准确性、可靠性或可用性的保证。</p>
              <p><strong>3.2 使用风险</strong><br/>用户理解并同意，使用本软件的风险完全由用户自行承担。作者和开发团队不对以下情况承担任何责任：本软件能否满足用户的特定需求；本软件的运行是否不间断、及时、安全或无错误；通过本软件获得的任何信息或服务是否准确可靠；本软件中任何缺陷是否会被纠正。</p>
              <p><strong>3.3 责任限制</strong><br/>在适用法律允许的最大范围内，作者和开发团队在任何情况下均不对因使用或无法使用本软件而产生的任何直接、间接、附带、特殊、惩罚性或后果性损害承担责任，包括但不限于：数据丢失或损坏；利润损失、业务中断；商誉损失；个人信息泄露；任何第三方索赔；因违反当地法律而产生的法律责任。</p>
              <p className={cn(
                'p-3 rounded-lg font-medium',
                themeStyle === 'apple-glass' ? 'bg-red-50 text-red-700' : 'bg-red-500/20 text-red-400'
              )}>
                <strong>3.4 用户行为责任</strong><br/>用户使用本软件的一切行为及其产生的任何后果，均由用户本人独立承担全部法律责任。任何因用户违反法律法规或本协议条款而产生的民事、行政或刑事责任，均与本软件作者和开发团队无关。
              </p>
              <p><strong>3.5 第三方内容</strong><br/>本软件可能包含指向第三方网站、服务或内容的链接或集成。作者和开发团队不对任何第三方内容、隐私政策或做法负责，也不对其进行认可或保证。</p>
            </>
          ) : (
            <>
              <p><strong>3.1 "As Is" Provision</strong><br/>This software is provided on an "AS IS" and "AS AVAILABLE" basis without any express or implied warranties, including but not limited to warranties of merchantability, fitness for a particular purpose, non-infringement, accuracy, reliability, or availability.</p>
              <p><strong>3.2 Use at Your Own Risk</strong><br/>Users understand and agree that the risk of using this software is entirely borne by the user. The author and development team are not responsible for: whether this software can meet the user's specific needs; whether the operation of this software is uninterrupted, timely, secure, or error-free; whether any information or services obtained through this software are accurate and reliable; whether any defects in this software will be corrected.</p>
              <p><strong>3.3 Limitation of Liability</strong><br/>To the maximum extent permitted by applicable law, the author and development team shall not be liable for any direct, indirect, incidental, special, punitive, or consequential damages arising from the use or inability to use this software, including but not limited to: data loss or damage; loss of profits, business interruption; loss of goodwill; personal information leakage; any third-party claims; legal liability arising from violations of local laws.</p>
              <p className={cn(
                'p-3 rounded-lg font-medium',
                themeStyle === 'apple-glass' ? 'bg-red-50 text-red-700' : 'bg-red-500/20 text-red-400'
              )}>
                <strong>3.4 User Conduct Responsibility</strong><br/>All actions taken by users using this software and any consequences arising therefrom shall be the sole legal responsibility of the user. Any civil, administrative, or criminal liability arising from the user's violation of laws, regulations, or the terms of this agreement has nothing to do with the author and development team of this software.
              </p>
              <p><strong>3.5 Third-Party Content</strong><br/>This software may contain links or integrations to third-party websites, services, or content. The author and development team are not responsible for any third-party content, privacy policies, or practices, nor do they endorse or guarantee them.</p>
            </>
          )}
        </div>
      </div>

      {/* Section 4: Privacy Policy */}
      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="app-icon green">
            <Lock className="w-4 h-4" />
          </div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {isZh ? '第四条 隐私政策' : 'Article 4: Privacy Policy'}
          </h2>
        </div>
        <div className={cn(
          'space-y-3 text-sm leading-relaxed',
          themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
        )}>
          {isZh ? (
            <>
              <p><strong>4.1 数据收集</strong><br/>本软件承诺不主动收集、存储、传输或分享任何用户个人身份信息（PII）。本软件不包含任何形式的用户追踪、行为分析或广告投放组件。</p>
              <p><strong>4.2 本地存储</strong><br/>本软件运行过程中产生的所有配置文件、日志文件、缓存数据等均存储在用户本地设备上，由用户自行管理和控制。</p>
              <p><strong>4.3 网络通信</strong><br/>本软件在运行过程中可能产生的网络通信包括：订阅源更新（用户主动配置）；规则集更新（用户主动配置）；版本检查（可禁用）；用户配置的代理连接。上述所有网络通信均基于用户的主动配置和操作，本软件不会在用户不知情的情况下进行任何网络通信。</p>
              <p><strong>4.4 日志记录</strong><br/>本软件的日志功能仅用于帮助用户排查技术问题，日志内容仅存储在本地，不会上传至任何服务器。用户可随时清除日志数据。</p>
            </>
          ) : (
            <>
              <p><strong>4.1 Data Collection</strong><br/>This software promises not to actively collect, store, transmit, or share any user personally identifiable information (PII). This software does not contain any form of user tracking, behavior analysis, or advertising components.</p>
              <p><strong>4.2 Local Storage</strong><br/>All configuration files, log files, cache data, etc. generated during the operation of this software are stored on the user's local device and are managed and controlled by the user.</p>
              <p><strong>4.3 Network Communication</strong><br/>Network communications that may occur during the operation of this software include: subscription source updates (actively configured by users); rule set updates (actively configured by users); version checking (can be disabled); proxy connections configured by users. All the above network communications are based on the user's active configuration and operation. This software will not conduct any network communication without the user's knowledge.</p>
              <p><strong>4.4 Logging</strong><br/>The logging function of this software is only used to help users troubleshoot technical problems. Log content is only stored locally and will not be uploaded to any server. Users can clear log data at any time.</p>
            </>
          )}
        </div>
      </div>

      {/* Section 5-8: Other Terms */}
      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="app-icon purple">
            <ScrollText className="w-4 h-4" />
          </div>
          <h2 className={cn(
            'text-lg font-semibold',
            themeStyle === 'apple-glass' ? 'text-slate-800' : 'text-white'
          )}>
            {isZh ? '其他条款' : 'Other Terms'}
          </h2>
        </div>
        <div className={cn(
          'space-y-4 text-sm leading-relaxed',
          themeStyle === 'apple-glass' ? 'text-slate-600' : 'text-slate-300'
        )}>
          {isZh ? (
            <>
              <div>
                <p className="font-medium mb-2">第五条 知识产权声明</p>
                <p>本软件的名称、标识、界面设计、源代码等知识产权归开发团队及相关贡献者所有。本软件中使用的第三方开源组件，其知识产权归各自的权利人所有。未经书面授权，任何个人或组织不得将本软件或其任何部分用于商业目的。</p>
              </div>
              <div>
                <p className="font-medium mb-2">第六条 服务变更与终止</p>
                <p>作者和开发团队保留随时修改、暂停或终止本软件或其任何部分的权利，恕不另行通知。本协议条款可能会不时更新，继续使用本软件即表示您接受更新后的条款。用户可随时通过卸载本软件终止使用。</p>
              </div>
              <div>
                <p className="font-medium mb-2">第七条 争议解决</p>
                <p>本协议的解释、效力及争议解决均适用中华人民共和国法律（不包括其冲突法规则）。因本协议或本软件使用产生的任何争议，各方应首先通过友好协商解决。</p>
              </div>
              <div>
                <p className="font-medium mb-2">第八条 其他</p>
                <p>本协议构成用户与作者之间关于使用本软件的完整协议。如本协议的任何条款被认定为无效或不可执行，不影响本协议其他条款的效力。</p>
              </div>
            </>
          ) : (
            <>
              <div>
                <p className="font-medium mb-2">Article 5: Intellectual Property</p>
                <p>The intellectual property rights of this software's name, logo, interface design, source code, etc. belong to the development team and related contributors. The intellectual property rights of third-party open-source components used in this software belong to their respective owners. Without written authorization, no individual or organization may use this software or any part of it for commercial purposes.</p>
              </div>
              <div>
                <p className="font-medium mb-2">Article 6: Service Changes & Termination</p>
                <p>The author and development team reserve the right to modify, suspend, or terminate this software or any part of it at any time without notice. The terms of this agreement may be updated from time to time. Continued use of this software indicates acceptance of the updated terms. Users may terminate use at any time by uninstalling this software.</p>
              </div>
              <div>
                <p className="font-medium mb-2">Article 7: Dispute Resolution</p>
                <p>The interpretation, validity, and dispute resolution of this agreement shall be governed by the laws of the People's Republic of China (excluding its conflict of law rules). Any disputes arising from this agreement or the use of this software should first be resolved through friendly negotiation.</p>
              </div>
              <div>
                <p className="font-medium mb-2">Article 8: Miscellaneous</p>
                <p>This agreement constitutes the entire agreement between the user and the author regarding the use of this software. If any provision of this agreement is found to be invalid or unenforceable, it shall not affect the validity of other provisions of this agreement.</p>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Final Notice */}
      <div className={cn(
        'glass-card p-6 border-2',
        themeStyle === 'apple-glass' ? 'border-red-200 bg-red-50/50' : 'border-red-500/30 bg-red-500/10'
      )}>
        <div className="flex items-start gap-4">
          <AlertTriangle className="w-8 h-8 text-red-500 flex-shrink-0" />
          <div>
            <h3 className={cn(
              'text-lg font-bold mb-3',
              themeStyle === 'apple-glass' ? 'text-red-700' : 'text-red-400'
            )}>
              {isZh ? '特别声明' : 'Special Declaration'}
            </h3>
            <div className={cn(
              'space-y-3 text-sm leading-relaxed',
              themeStyle === 'apple-glass' ? 'text-slate-700' : 'text-slate-300'
            )}>
              {isZh ? (
                <>
                  <p>本软件是一款技术工具，工具本身不具有违法性。但任何工具都可能被滥用。<strong>请务必确保您的使用行为符合当地法律法规。</strong></p>
                  <p className="font-bold">对于任何用户因违法使用本软件而导致的法律后果，本软件作者和开发团队概不负责，一切法律责任由违法行为人自行承担。</p>
                  <p>如果您所在的国家或地区禁止使用此类软件，请立即停止使用并删除本软件。</p>
                </>
              ) : (
                <>
                  <p>This software is a technical tool, and the tool itself is not illegal. However, any tool can be misused. <strong>Please make sure your use complies with local laws and regulations.</strong></p>
                  <p className="font-bold">The author and development team of this software are not responsible for any legal consequences caused by users' illegal use of this software. All legal responsibilities shall be borne by the violators themselves.</p>
                  <p>If your country or region prohibits the use of such software, please stop using it immediately and delete this software.</p>
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Agreement */}
      <div className={cn(
        'text-center py-6 text-sm',
        themeStyle === 'apple-glass' ? 'text-slate-500' : 'text-slate-400'
      )}>
        {isZh 
          ? '使用本软件即表示您已完整阅读、充分理解并同意接受本协议的全部条款。'
          : 'By using this software, you acknowledge that you have fully read, understood, and agreed to all terms of this agreement.'
        }
      </div>
    </div>
  )
}
