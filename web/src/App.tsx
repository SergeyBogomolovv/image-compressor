import { useState } from 'react'

interface UploadResponse {
  success: string[]
  errors: string[]
}

const App = () => {
  const [successLinks, setSuccessLinks] = useState<string[]>([])
  const [errorCount, setErrorCount] = useState<number>(0)
  const serverUrl = import.meta.env.VITE_SERVER_URL

  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    if (!event.target.files) return

    const formData = new FormData()
    Array.from(event.target.files).forEach((file) => {
      formData.append('images', file)
    })

    try {
      const response = await fetch(`${serverUrl}/upload`, {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        throw new Error('Upload failed')
      }

      const data: UploadResponse = await response.json()
      setSuccessLinks(data.success)
      setErrorCount(data.errors.length)
    } catch (error) {
      setErrorCount(1)
    }
  }

  return (
    <div className='container'>
      <h1>Image Compressor</h1>
      <p>
        Сервис сжатия изображения для ваших сайтов и приложений. Загрузите изображения и получите
        уникальные ссылки на сжатые версии.
      </p>
      <div className='upload-section'>
        <label htmlFor='file-upload' className='upload-button'>
          Загрузить
        </label>
        <input id='file-upload' type='file' multiple onChange={handleFileUpload} hidden />
      </div>
      <div className='result-section'>
        {successLinks.length > 0 && (
          <>
            <h2>Ссылки для скачивания</h2>
            <ul>
              {successLinks.map((link, index) => (
                <li key={index}>
                  <a href={`${serverUrl}/download/${link}`} download>
                    {link}
                  </a>
                </li>
              ))}
            </ul>
          </>
        )}
        {errorCount > 0 && <h3>Произошли ошибки: {errorCount}</h3>}
      </div>
    </div>
  )
}

export default App
