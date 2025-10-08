# OpenCV模型文件

## 📋 需要的模型文件

### 1. Haar Cascade人脸检测模型

**文件名**: `haarcascade_frontalface_default.xml`

**用途**: 人脸检测，用于美颜功能

**获取方式**:

#### 方法1: 从vcpkg安装的OpenCV复制

```powershell
# 复制文件
Copy-Item "C:\vcpkg\installed\x64-windows\share\opencv4\haarcascades\haarcascade_frontalface_default.xml" `
          "resources\models\"
```

#### 方法2: 从OpenCV GitHub下载

```powershell
# 下载文件
$url = "https://raw.githubusercontent.com/opencv/opencv/master/data/haarcascades/haarcascade_frontalface_default.xml"
Invoke-WebRequest -Uri $url -OutFile "resources\models\haarcascade_frontalface_default.xml"
```

#### 方法3: 从系统OpenCV目录复制

如果您安装了OpenCV到其他位置，可以从以下路径复制：
- `C:\opencv\build\etc\haarcascades\haarcascade_frontalface_default.xml`
- `C:\OpenCV\sources\data\haarcascades\haarcascade_frontalface_default.xml`

---

### 2. DNN人像分割模型 (可选)

**文件名**: 
- `frozen_inference_graph.pb`
- `ssd_mobilenet_v2_coco.pbtxt`

**用途**: 高精度人像分割，用于虚拟背景功能

**获取方式**:

#### 从TensorFlow Model Zoo下载

```powershell
# 下载SSD MobileNet V2模型
$modelUrl = "http://download.tensorflow.org/models/object_detection/ssd_mobilenet_v2_coco_2018_03_29.tar.gz"
Invoke-WebRequest -Uri $modelUrl -OutFile "ssd_mobilenet_v2.tar.gz"

# 解压并复制到resources/models/
```

**注意**: 这些模型是可选的。如果没有这些模型，程序会使用Background Subtraction MOG2作为替代方案。

---

## 🔧 安装脚本

### 自动安装脚本

创建 `install-models.ps1`:

```powershell
# 创建目录
New-Item -ItemType Directory -Path "resources\models" -Force

# 下载Haar Cascade模型
Write-Host "Downloading Haar Cascade model..." -ForegroundColor Cyan
$url = "https://raw.githubusercontent.com/opencv/opencv/master/data/haarcascades/haarcascade_frontalface_default.xml"
Invoke-WebRequest -Uri $url -OutFile "resources\models\haarcascade_frontalface_default.xml"

Write-Host "Model downloaded successfully!" -ForegroundColor Green
```

### 从vcpkg复制脚本

创建 `copy-models-from-vcpkg.ps1`:

```powershell
$vcpkgPath = "C:\vcpkg\installed\x64-windows\share\opencv4\haarcascades"
$targetPath = "resources\models"

# 创建目录
New-Item -ItemType Directory -Path $targetPath -Force

# 复制文件
if (Test-Path "$vcpkgPath\haarcascade_frontalface_default.xml") {
    Copy-Item "$vcpkgPath\haarcascade_frontalface_default.xml" $targetPath
    Write-Host "Model copied successfully!" -ForegroundColor Green
} else {
    Write-Host "Model not found in vcpkg!" -ForegroundColor Red
}
```

---

## 📝 添加到资源文件

在 `resources/resources.qrc` 中添加:

```xml
<qresource prefix="/models">
    <file>models/haarcascade_frontalface_default.xml</file>
</qresource>
```

---

## ✅ 验证安装

### 检查文件是否存在

```powershell
# 检查Haar Cascade模型
Test-Path "resources\models\haarcascade_frontalface_default.xml"
```

### 检查文件大小

```powershell
# 应该约为930KB
(Get-Item "resources\models\haarcascade_frontalface_default.xml").Length / 1KB
```

---

## 🎯 使用说明

### 在代码中使用

模型文件会被编译到可执行文件中，可以通过Qt资源系统访问：

```cpp
// 加载Haar Cascade模型
QString cascadePath = ":/models/haarcascade_frontalface_default.xml";
if (!m_faceCascade.load(cascadePath.toStdString())) {
    qWarning() << "Failed to load face cascade";
}
```

### 回退机制

如果资源文件中的模型加载失败，程序会尝试从系统路径加载：

```cpp
// 尝试系统OpenCV数据路径
std::string systemPath = cv::samples::findFile("haarcascade_frontalface_default.xml");
if (!m_faceCascade.load(systemPath)) {
    qWarning() << "Failed to load face cascade classifier";
}
```

---

## 📊 模型文件信息

| 文件 | 大小 | 必需 | 用途 |
|------|------|------|------|
| haarcascade_frontalface_default.xml | ~930KB | ✅ 是 | 人脸检测 |
| frozen_inference_graph.pb | ~67MB | ❌ 否 | 人像分割 |
| ssd_mobilenet_v2_coco.pbtxt | ~77KB | ❌ 否 | 模型配置 |

---

## ❓ 常见问题

### Q: 必须安装所有模型吗？

**A**: 不是。只有 `haarcascade_frontalface_default.xml` 是必需的。DNN模型是可选的，用于提高人像分割精度。

### Q: 模型文件很大，会影响程序大小吗？

**A**: 
- Haar Cascade模型只有930KB，影响很小
- DNN模型较大(67MB)，建议作为可选下载

### Q: 可以使用其他模型吗？

**A**: 可以。您可以使用任何OpenCV兼容的模型，只需修改代码中的路径即可。

### Q: 模型加载失败怎么办？

**A**: 
1. 检查文件是否存在
2. 检查文件路径是否正确
3. 检查resources.qrc是否包含该文件
4. 重新编译项目

---

## 🔗 相关链接

- [OpenCV Haar Cascades](https://github.com/opencv/opencv/tree/master/data/haarcascades)
- [TensorFlow Model Zoo](https://github.com/tensorflow/models/blob/master/research/object_detection/g3doc/tf1_detection_zoo.md)
- [OpenCV DNN Module](https://docs.opencv.org/master/d2/d58/tutorial_table_of_content_dnn.html)

---

**📝 注意**: 请确保在编译项目前下载并放置好所需的模型文件。

