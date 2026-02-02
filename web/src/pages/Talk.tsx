import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../utils/api'
import { getUser, logout, isAdmin } from '../utils/auth'

export default function Talk() {
  const [isRecording, setIsRecording] = useState(false)
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const navigate = useNavigate()
  const user = getUser()

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mediaRecorder = new MediaRecorder(stream)
      mediaRecorderRef.current = mediaRecorder
      chunksRef.current = []

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data)
        }
      }

      mediaRecorder.onstop = async () => {
        const audioBlob = new Blob(chunksRef.current, { type: 'audio/webm' })
        await uploadAudio(audioBlob)
        stream.getTracks().forEach((track) => track.stop())
      }

      mediaRecorder.start()
      setIsRecording(true)
      setMessage('')
      setMessageType('')
    } catch (err) {
      showMessage('无法访问麦克风', 'error')
    }
  }

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop()
      setIsRecording(false)
    }
  }

  const uploadAudio = async (audioBlob: Blob) => {
    const formData = new FormData()
    formData.append('audio', audioBlob, 'recording.webm')

    try {
      const response = await api.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      showMessage(response.data.text || '上传成功', 'success')
    } catch (err: any) {
      showMessage(err.response?.data?.error || '上传失败', 'error')
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
              onMouseLeave={stopRecording}
              onTouchStart={startRecording}
              onTouchEnd={stopRecording}
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
            <p>🖱️ 鼠标按住录音，松开发送</p>
            <p>📱 触摸屏按住录音，松开发送</p>
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
            <li>• 按住录音按钮开始录音</li>
            <li>• 松开按钮自动上传并发送到语音识别服务</li>
            <li>• 录音时会显示红色状态</li>
            <li>• 确保浏览器已授权麦克风权限</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
