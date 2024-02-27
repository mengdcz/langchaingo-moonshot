package params

// 配额/速率，默认情况下，每个 API 密钥对 API 的请求限制为每分钟 60 个。如果您想要更高的值，请告诉我们。

// 文生图
type ImagesRequest struct {
	Prompt string `json:"prompt"` // 必须 最大长度为 1000 个字符
}
type ImagesResponse struct {
	Success           bool   `json:"success"`
	ImgFile           []byte `json:"img_file,omitempty"`            // 成功响应 mime-type 为image/png 正文为尺寸为 1024x1024 像素的图像
	XRemainingCredits string `json:"x_remaining_credits,omitempty"` // 还剩下多少积分
	XReditsConsumed   string `json:"x_redits_consumed,omitempty"`   // 请求消耗了多少积分
	ImgExt            string `json:"img_ext,omitempty"`             // 图片后缀
	Error             string `json:"error,omitempty"`               // 错误类型由响应状态码指示，详细信息在 json 正文中
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"` // 错误类型由响应状态码指示，详细信息在 json 正文中
}

// 取消裁剪
type UnCropRequest struct {
	ImageFile   string `json:"image_file"`             // 必须 原始图像应为 JPG、PNG 或 WebP，最大分辨率为 10 兆像素，最大文件大小为 30 Mb。
	ExtendLeft  int64  `json:"extend_left,omitempty"`  // 可选 最大为 2k，默认为 0 【正负2k】
	ExtendRight int64  `json:"extend_right,omitempty"` // 可选
	ExtendUp    int64  `json:"extend_up,omitempty"`    // 可选
	ExtendDown  int64  `json:"extend_down,omitempty"`  // 可选
}
type UnCropResponse struct {
	ImagesResponse // 成功响应 mime-type 是image/jpeg，响应图像是未裁剪的 JPEG 图像
}

// 文本修复
type TextInpaintingRequest struct {
	ImageFile  string `json:"image_file"`  // 必须 要处理的原始图像, 应为 JPG 或 PNG，最大分辨率为 10 兆像素，最大文件大小为 30 Mb
	MaskFile   string `json:"mask_file"`   // 必须 掩模图像，定义需要删除的区域。 遮罩图像应为 PNG，并且应具有与原始图像相同的分辨率，最大文件大小为 30 Mb,蒙版应该是黑白的，没有灰色像素（例如，值仅为 0 或 255），值 0 表示保持原样的像素，255 表示要“替换”的像素
	TextPrompt string `json:"text_prompt"` // 必须 描述您要在图像中放入的内容
}
type TextInpaintingResponse struct {
	ImagesResponse // 成功响应 mime-type 是image/jpeg，响应图像是与 尺寸相同的 JPEG 图像
}

// 草图到图像
type SketchToImageRequest struct {
	SketchFile string `json:"sketch_file"` // 必须 草图, 草图应为 PNG、JPEG 或 WebP 文件中黑色背景上的白色草图线，最大宽度和高度为 1024 像素。目前仅支持方形图像
	Prompt     string `json:"prompt"`      // 必须 描述要生成的内容, 最大长度为 5000 个字符
}
type SketchToImageResponse struct {
	ImagesResponse // 成功响应 mime-type 是image/jpeg，正文将包含生成的 jpeg 格式的图像
}

// 替换背景
type ReplaceBackgroundRequest struct {
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 输入图像应为 PNG、JPG 或 WEBP 文件，最大宽度和高度为 2048 像素，最大文件大小为 20 Mb
	Prompt    string `json:"prompt"`     // 必须 描述您要将项目传送到的场景。该字段的值可以是空字符串，在这种情况下，我们将根据您的项目生成场景。
}
type ReplaceBackgroundResponse struct {
	ImagesResponse // 成功响应 mime-type 将为image/png,image/webp或image/jpeg（请参阅格式部分），响应正文将是相应格式的图像，其尺寸与输入图像相同
}

// 删除文本
type RemoveTextRequest struct {
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 原始图像应为 JPG 或 PNG，最大分辨率为 16 兆像素，最大文件大小为 30 Mb
}
type RemoveTextResponse struct {
	ImagesResponse // 成功响应 mime-type  是image/png，响应图像是与 尺寸相同的 PNG 图像image_file
}

// 删除背景
type RemoveBackgroundRequest struct {
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 原始图像应为 PNG、JPG 或 WEBP 文件，最大分辨率为 25 兆像素，最大文件大小为 30 Mb。
}
type RemoveBackgroundResponse struct {
	ImagesResponse // 成功响应 mime-type 将为image/png、image/webp或image/jpeg，并且响应图像将是与接受标头匹配的图像（如果提供），否则为 PNG，与 具有相同的尺寸image_file
}

// 重新想象
type ReimagineRequest struct {
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 原始图像应为 PNG、JPEG 或 WebP 文件，最大宽度和高度为 1024 像素。
}
type ReImagesResponse struct {
	ImagesResponse // 成功响应 mime-type image/jpeg，响应正文将包含 jpeg 格式的重新想象的图像
}

// 肖像表面法线
type PortraitSurfaceNormalsRequest struct {
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 原始图像应为 PNG、JPEG 或 WEBP 文件，最大分辨率为 25 兆像素，最大文件大小为 30 Mb。
}
type PortraitSurfaceNormalsResponse struct {
	ImagesResponse // 成功响应 mime-type image/jpeg，响应正文将包含 jpeg 响应主体将包含一个图像，该图像包含输入图像的表面法线。
}

// 人像深度估计
type PortraitDepthEstimationRequest struct { //  专注于肖像深度估计，因此任何不属于人物的图像像素都将被忽略（即，这些像素的深度估计将为 0）
	ImageFile string `json:"image_file"` // 必须 要处理的原始图像, 原始图像应为 PNG、JPEG 或 WEBP 文件，最大分辨率为 25 兆像素，最大文件大小为 30 Mb。
}
type PortraitDepthEstimationResponse struct {
	ImagesResponse // 成功响应 mime-type image/jpeg，结果图像将是 JPEG，其中每个像素值是该像素的深度估计，白色代表最近的像素，黑色代表最远的像素。
}

// 图像放大-同步接口【支持异步：1、注册您的图像放大，2、检查任务（每 5 秒左右轮询一次），3、获取放大】
type ImageUpscaleRequest struct {
	ImageFile    string `json:"image_file"`    // 必须 要处理的原始图像,原始图像应为 PNG、JPEG 或 WebP 文件，最大分辨率为 16 兆像素，最大文件大小为 30 Mb
	TargetWidth  int64  `json:"target_width"`  // 必须 所需的像素宽度，同步：有效值为 1 到 4096 之间的整数。异步：必须介于 1 和 16384 之间
	TargetHeight int64  `json:"target_height"` // 必须 所需的像素高度，同步：有效值为 1 到 4096 之间的整数。异步：必须介于 1 和 16384 之间
}
type ImageUpscaleResponse struct {
	ImagesResponse // 成功响应 mime 类型将为 image/ [webp/jpeg]，如果原始图像包含透明度，则响应正文将包含 webp 格式的放大图像，如果不包含透明度，则响应正文将包含 jpeg 格式。
}

// 清理
type CleanupRequest struct {
	ImageFile string `json:"image_file"`                          // 必须 要处理的原始图像,原始图像应为 PNG、JPEG 或 WebP 文件，最大分辨率为 16 兆像素，最大文件大小为 30 Mb
	MaskFile  string `json:"mask_file"`                           // 必须 要处理的原始图像,原始图像应为 PNG、JPEG 或 WebP 文件，最大分辨率为 16 兆像素，最大文件大小为 30 Mb
	Mode      string `json:"mode,omitempty,options=fast|quality"` // 可选字段，可以设置为fast或quality控制速度和质量之间的权衡, fast是默认模式，速度更快，但可能会在结果图像中产生伪影; quality速度较慢，但会产生更好的结果
}
type CleanupResponse struct {
	ImagesResponse // 成功响应 mime-type 是image/png，响应图像是与 尺寸相同的 PNG 图像image_file
}
