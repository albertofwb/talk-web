import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../utils/api'
import { getUser, logout, isAdmin } from '../utils/auth'

export default function Talk() {
  const [isRecording, setIsRecording] = useState(false)
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')
  const [micPermission, setMicPermission] = useState<'prompt' | 'granted' | 'denied'>('prompt')
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const recordingStartTimeRef = useRef<number>(0)
  const navigate = useNavigate()
  const user = getUser()

  const MIN_RECORDING_TIME = 500 // 最小录音时长（毫秒）

  // 组件卸载时清理麦克风流
  useEffect(() => {
    return () => {
      if (streamRef.current) {
        streamRef.current.getTracks().forEach(track => track.stop())
        streamRef.current = null
      }
    }
  }, [])

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

  const startRecording = async () => {
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
    }
  }

  const stopRecording = () => {
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
    const formData = new FormData()
    formData.append('audio', audioBlob, 'recording.webm')

    try {
      showMessage('识别中...', 'success')

      const response = await api.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })

      const { text, reply, reply_audio, tts_error } = response.data

      // 显示识别的文字
      if (text) {
        showMessage(`✓ ${text}`, 'success')
      } else {
        showMessage('未识别到语音内容', 'error')
        return
      }

      // 播放回复语音
      if (reply_audio) {
        const audio = new Audio(reply_audio)
        audio.play().catch(err => {
          console.error('播放音频失败:', err)
        })
      } else if (tts_error) {
        console.warn('TTS生成失败:', tts_error)
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.detail || err.response?.data?.error || err.message || '上传失败'
      showMessage(`❌ ${errorMsg}`, 'error')
      console.error('上传错误:', err.response?.data)
    }
  }

  const showMessage = (text: string, type: 'success' | 'error') => {
    setMessage(text)
    setMessageType(type)
    setTimeout(() => {
      setMessage('')
      setMessageType('')
    }, 3000)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-pink-50">
      {/* 顶部导航 */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div>
            <h1 className="text-xl font-bold text-gray-800">语音对讲</h1>
            <p className="text-sm text-gray-500">欢迎, {user?.username}</p>
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
              onTouchStart={startRecording}
              onTouchEnd={stopRecording}
              onTouchCancel={stopRecording}
              className={`w-48 h-48 rounded-full text-white font-bold text-xl shadow-2xl transition-all duration-200 ${
                isRecording
                  ? 'bg-red-500 scale-110'
                  : 'bg-indigo-600 hover:bg-indigo-700 active:scale-95'
              }`}
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
