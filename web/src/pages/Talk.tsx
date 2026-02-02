import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../utils/api'
import { getUser, logout, isAdmin } from '../utils/auth'

interface HistoryMessage {
  id: number
  text: string
  reply: string
  status: string
  sent_at: string
  replied_at?: string
}

export default function Talk() {
  const [isRecording, setIsRecording] = useState(false)
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')
  const [_micPermission, setMicPermission] = useState<'prompt' | 'granted' | 'denied'>('prompt')
  const [history, setHistory] = useState<HistoryMessage[]>([])
  const [wsConnected, setWsConnected] = useState(false)
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const recordingStartTimeRef = useRef<number>(0)
  const wsRef = useRef<WebSocket | null>(null)
  const wsReconnectTimerRef = useRef<number | null>(null)
  const wsReconnectAttemptsRef = useRef<number>(0)
  const navigate = useNavigate()
  const user = getUser()

  const MIN_RECORDING_TIME = 500 // 最小录音时长（毫秒）
  const WS_RECONNECT_DELAYS = [1000, 2000, 5000, 10000, 30000] // 重连延迟（递增）
  const isMouseDownRef = useRef(false) // 跟踪鼠标是否按下

  // 加载历史记录
  const loadHistory = async (playLatestAudio = false) => {
    try {
      const response = await api.get('/history')
      const messages = response.data.messages || []
      setHistory(messages)

      // 如果需要播放最新音频
      if (playLatestAudio && messages.length > 0) {
        const latestMsg = messages[0] // 最新的消息（按时间倒序）
        if (latestMsg.reply_audio) {
          console.log('自动播放最新音频:', latestMsg.reply_audio)
          playAudio(latestMsg.reply_audio)
        }
      }
    } catch (err) {
      console.error('加载历史失败:', err)
    }
  }

  // 建立 WebSocket 连接（带重连策略）
  const connectWebSocket = () => {
    const token = localStorage.getItem('token')
    if (!token) return

    // 清除之前的重连定时器
    if (wsReconnectTimerRef.current) {
      clearTimeout(wsReconnectTimerRef.current)
      wsReconnectTimerRef.current = null
    }

    // 使用当前页面的 host（适配开发和生产环境）
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/ws?token=${token}`

    console.log('连接 WebSocket:', wsUrl, `(尝试 ${wsReconnectAttemptsRef.current + 1})`)
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket 已连接')
      setWsConnected(true)
      wsReconnectAttemptsRef.current = 0 // 重置重连计数
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('收到 WebSocket 消息:', data)

        if (data.type === 'reply') {
          const { reply, reply_audio } = data.data

          // 显示回复
          showMessage(`💬 ${reply}`, 'success')

          // 刷新历史并播放音频
          // 如果 WebSocket 消息中有 reply_audio，直接播放
          // 否则从历史记录中获取最新的音频播放
          if (reply_audio) {
            playAudio(reply_audio)
            loadHistory() // 刷新历史但不播放
          } else {
            loadHistory(true) // 刷新历史并自动播放最新音频
          }
        }
      } catch (err) {
        console.error('解析消息失败:', err)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket 错误:', error)
      setWsConnected(false)
    }

    ws.onclose = () => {
      console.log('WebSocket 已断开')
      setWsConnected(false)

      // 计算重连延迟（递增退避策略）
      const delayIndex = Math.min(wsReconnectAttemptsRef.current, WS_RECONNECT_DELAYS.length - 1)
      const delay = WS_RECONNECT_DELAYS[delayIndex]

      console.log(`${delay / 1000}秒后尝试重连...`)
      wsReconnectAttemptsRef.current++

      wsReconnectTimerRef.current = window.setTimeout(() => {
        connectWebSocket()
      }, delay)
    }

    wsRef.current = ws
  }

  // 组件挂载时加载历史并建立 WebSocket
  useEffect(() => {
    loadHistory()
    connectWebSocket()

    return () => {
      // 清理 WebSocket 连接
      if (wsRef.current) {
        wsRef.current.close()
      }
      // 清理重连定时器
      if (wsReconnectTimerRef.current) {
        clearTimeout(wsReconnectTimerRef.current)
      }
    }
  }, [])

  // 组件卸载时清理麦克风流
  useEffect(() => {
    return () => {
      if (streamRef.current) {
        streamRef.current.getTracks().forEach(track => track.stop())
        streamRef.current = null
      }
    }
  }, [])

  // 全局 mouseup 监听（防止鼠标事件被打断）
  useEffect(() => {
    const handleGlobalMouseUp = () => {
      if (isMouseDownRef.current && isRecording) {
        console.log('全局 mouseup 触发，停止录音')
        stopRecording()
      }
    }

    document.addEventListener('mouseup', handleGlobalMouseUp)
    return () => {
      document.removeEventListener('mouseup', handleGlobalMouseUp)
    }
  }, [isRecording])

  // 初始化麦克风（只请求一次权限）
  const initMicrophone = async () => {
    if (streamRef.current) {
      return streamRef.current
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      streamRef.current = stream
      setMicPermission('granted')
      return stream
    } catch (err: any) {
      setMicPermission('denied')
      if (err.name === 'NotAllowedError') {
        showMessage('麦克风权限被拒绝，请在浏览器设置中允许', 'error')
      } else {
        showMessage(`麦克风错误: ${err.message}`, 'error')
      }
      throw err
    }
  }

  const startRecording = async (e: React.MouseEvent | React.TouchEvent) => {
    // 只阻止文本选择，不阻止其他默认行为
    if (e.type === 'mousedown') {
      e.preventDefault()
      isMouseDownRef.current = true
    }

    // 防止重复启动
    if (isRecording) return

    try {
      // 获取或初始化麦克风流
      const stream = await initMicrophone()

      // 尝试使用 opus 编码的 webm，如果不支持则使用默认
      let options = { mimeType: 'audio/webm;codecs=opus' }
      if (!MediaRecorder.isTypeSupported(options.mimeType)) {
        options = { mimeType: 'audio/webm' }
      }

      const mediaRecorder = new MediaRecorder(stream, options)
      mediaRecorderRef.current = mediaRecorder
      chunksRef.current = []

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data)
        }
      }

      mediaRecorder.onstop = async () => {
        const recordingDuration = Date.now() - recordingStartTimeRef.current

        // 检查录音时长
        if (recordingDuration < MIN_RECORDING_TIME) {
          showMessage('录音时间太短，请按住至少1秒', 'error')
          return
        }

        // 等待一下确保数据收集完成
        await new Promise(resolve => setTimeout(resolve, 100))

        const audioBlob = new Blob(chunksRef.current, { type: 'audio/webm;codecs=opus' })

        // 检查音频大小
        if (audioBlob.size < 1000) {
          showMessage('录音数据太小，请重试并说话', 'error')
          return
        }

        console.log(`录音完成: ${recordingDuration}ms, 大小: ${audioBlob.size} bytes`)
        await uploadAudio(audioBlob)
      }

      // 每 100ms 收集一次数据
      mediaRecorder.start(100)
      recordingStartTimeRef.current = Date.now()
      setIsRecording(true)
      setMessage('')
      setMessageType('')
    } catch (err: any) {
      // 错误已在 initMicrophone 中处理
      setIsRecording(false)
    }
  }

  const stopRecording = (e?: React.MouseEvent | React.TouchEvent) => {
    // 只在鼠标事件时阻止默认行为
    if (e && e.type === 'mouseup') {
      e.preventDefault()
      isMouseDownRef.current = false
    }

    // 鼠标移出时也重置状态
    if (e && e.type === 'mouseleave') {
      isMouseDownRef.current = false
    }

    if (mediaRecorderRef.current && isRecording) {
      const recorder = mediaRecorderRef.current

      // 检查录音器状态
      if (recorder.state === 'recording') {
        recorder.stop()
      } else {
        console.warn('MediaRecorder 状态异常:', recorder.state)
      }

      setIsRecording(false)
    }
  }

  const uploadAudio = async (audioBlob: Blob) => {
    // 生成唯一消息ID (timestamp + random)
    const msgId = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    console.log('📤 [上传] 生成消息ID:', msgId)

    const formData = new FormData()
    formData.append('audio', audioBlob, 'recording.webm')
    formData.append('msg_id', msgId)  // 添加消息ID

    try {
      showMessage('识别中...', 'success')

      const response = await api.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })

      const { text, message_id } = response.data
      console.log('✓ [上传] 后端返回 message_id:', message_id)

      // 显示识别的文字
      if (text) {
        if (wsConnected) {
          showMessage(`✓ ${text} (等待回复...)`, 'success')
          // WebSocket 会自动推送回复，无需轮询
        } else {
          showMessage(`✓ ${text} (等待回复...)`, 'success')
          // WebSocket 未连接，使用轮询兜底
          pollForReply()
        }
      } else {
        showMessage('未识别到语音内容', 'error')
        return
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.detail || err.response?.data?.error || err.message || '上传失败'
      showMessage(`❌ ${errorMsg}`, 'error')
      console.error('上传错误:', err.response?.data)
    }
  }

  const pollForReply = async () => {
    const maxAttempts = 60 // 最多轮询 60 次（60秒）
    let attempts = 0

    const poll = async () => {
      if (attempts >= maxAttempts) {
        showMessage('⏱️ 等待回复超时', 'error')
        return
      }

      attempts++

      try {
        const response = await api.get('/reply')
        const { status, reply, reply_audio } = response.data

        if (status === 'ready' && reply) {
          // 收到回复
          showMessage(`💬 ${reply}`, 'success')

          // 刷新历史记录
          loadHistory()

          // 播放 TTS 音频
          if (reply_audio) {
            playAudio(reply_audio)
          } else {
            console.log('收到回复但没有音频:', reply)
          }
          return
        }

        // 还在等待，继续轮询
        if (status === 'waiting') {
          setTimeout(poll, 1000) // 1秒后再次轮询
        }
      } catch (err: any) {
        console.error('轮询错误:', err)
        showMessage('获取回复失败', 'error')
      }
    }

    poll()
  }

  const showMessage = (text: string, type: 'success' | 'error') => {
    setMessage(text)
    setMessageType(type)
    setTimeout(() => {
      setMessage('')
      setMessageType('')
    }, 3000)
  }

  // 播放音频（带认证）
  const playAudio = async (audioUrl: string) => {
    try {
      console.log('🔊 [播放音频] 开始:', audioUrl)

      // 使用 fetch 下载音频（audioUrl 已包含 /api 前缀）
      const token = localStorage.getItem('token')

      console.log('📥 [播放音频] 下载中...')
      const response = await fetch(audioUrl, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const blob = await response.blob()
      console.log('✓ [播放音频] 下载完成，大小:', blob.size, 'bytes')

      // 创建 Blob URL
      const blobUrl = URL.createObjectURL(blob)
      console.log('✓ [播放音频] Blob URL 创建:', blobUrl)

      // 播放
      const audio = new Audio(blobUrl)
      audio.onended = () => {
        console.log('✓ [播放音频] 播放完成')
        URL.revokeObjectURL(blobUrl) // 清理 Blob URL
      }
      audio.onerror = (err) => {
        console.error('✗ [播放音频] 播放失败:', err)
        URL.revokeObjectURL(blobUrl)
      }

      console.log('▶️ [播放音频] 开始播放...')
      await audio.play()
      console.log('✓ [播放音频] 播放成功')
    } catch (err: any) {
      console.error('✗ [播放音频] 失败:', err)
      console.error('错误详情:', {
        name: err.name,
        message: err.message,
        code: err.code
      })
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-pink-50">
      {/* 顶部导航 */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div>
            <h1 className="text-xl font-bold text-gray-800">语音对讲</h1>
            <div className="flex items-center gap-3">
              <p className="text-sm text-gray-500">欢迎, {user?.username}</p>
              <span className={`text-xs px-2 py-1 rounded-full ${
                wsConnected
                  ? 'bg-green-100 text-green-700'
                  : 'bg-red-100 text-red-700'
              }`}>
                {wsConnected ? '🟢 已连接' : '🔴 未连接'}
              </span>
            </div>
          </div>
          <div className="flex gap-3">
            {isAdmin() && (
              <button
                onClick={() => navigate('/admin')}
                className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition"
              >
                管理后台
              </button>
            )}
            <button
              onClick={logout}
              className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition"
            >
              登出
            </button>
          </div>
        </div>
      </nav>

      {/* 主要内容 */}
      <div className="max-w-2xl mx-auto px-4 py-16">
        <div className="bg-white rounded-3xl shadow-2xl p-12">
          <h2 className="text-2xl font-bold text-center text-gray-800 mb-8">
            按住按钮开始录音
          </h2>

          {/* 录音按钮 */}
          <div className="flex justify-center mb-8">
            <button
              onMouseDown={startRecording}
              onMouseUp={stopRecording}
              onMouseLeave={stopRecording}
              onTouchStart={startRecording}
              onTouchEnd={stopRecording}
              onTouchCancel={stopRecording}
              className={`w-48 h-48 rounded-full text-white font-bold text-xl shadow-2xl transition-all duration-200 select-none ${
                isRecording
                  ? 'bg-red-500 scale-110'
                  : 'bg-indigo-600 hover:bg-indigo-700 active:scale-95'
              }`}
              style={{ userSelect: 'none', WebkitUserSelect: 'none' }}
            >
              {isRecording ? '🎤 录音中...' : '按住说话'}
            </button>
          </div>

          {/* 提示信息 */}
          <div className="text-center text-gray-600 space-y-2">
            <p>🖱️ 鼠标按住录音（至少1秒），松开发送</p>
            <p>📱 触摸屏按住录音（至少1秒），松开发送</p>
            <p className="text-sm text-gray-500">⚠️ 请确保说话清晰，环境安静</p>
          </div>

          {/* 消息提示 */}
          {message && (
            <div
              className={`mt-6 p-4 rounded-lg text-center font-medium ${
                messageType === 'success'
                  ? 'bg-green-50 text-green-700 border border-green-200'
                  : 'bg-red-50 text-red-700 border border-red-200'
              }`}
            >
              {message}
            </div>
          )}
        </div>

        {/* 对话历史 */}
        {history.length > 0 && (
          <div className="mt-8 bg-white rounded-2xl shadow-lg p-6">
            <h3 className="font-bold text-gray-800 mb-4">最近对话</h3>
            <div className="space-y-4">
              {history.slice().reverse().map((msg) => (
                <div key={msg.id} className="border-l-4 border-indigo-500 pl-4 py-2">
                  <div className="flex items-start gap-2">
                    <span className="text-gray-500 text-sm">你:</span>
                    <p className="text-gray-800">{msg.text}</p>
                  </div>
                  {msg.reply && (
                    <div className="flex items-start gap-2 mt-2">
                      <span className="text-indigo-600 text-sm">AI:</span>
                      <p className="text-gray-700">{msg.reply}</p>
                    </div>
                  )}
                  {!msg.reply && msg.status !== 'timeout' && (
                    <p className="text-gray-400 text-sm mt-2">⏳ 等待回复...</p>
                  )}
                  {msg.status === 'timeout' && !msg.reply && (
                    <p className="text-red-400 text-sm mt-2">⏱️ 回复超时</p>
                  )}
                  {msg.reply && msg.replied_at && (
                    <p className="text-green-600 text-xs mt-1">
                      ✓ {new Date(msg.replied_at).toLocaleString('zh-CN', {
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit'
                      })}
                    </p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 使用说明 */}
        <div className="mt-8 bg-white rounded-2xl shadow-lg p-6">
          <h3 className="font-bold text-gray-800 mb-4">使用说明</h3>
          <ul className="space-y-2 text-gray-600">
            <li>• <strong>按住录音按钮至少1秒</strong>开始录音</li>
            <li>• 松开按钮自动上传并识别语音内容</li>
            <li>• 录音时会显示红色状态</li>
            <li>• 识别成功后会自动播放回复语音</li>
            <li>• 确保浏览器已授权麦克风权限（HTTPS）</li>
            <li>• 环境安静，说话清晰效果更好</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
