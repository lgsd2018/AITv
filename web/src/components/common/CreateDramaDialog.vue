<template>
  <!-- Create Drama Dialog / 创建短剧弹窗 -->
  <el-dialog
    v-model="visible"
    :title="$t('drama.createNew')"
    width="600px"
    :close-on-click-modal="false"
    class="create-dialog"
    @closed="handleClosed"
  >
    <div class="dialog-desc">{{ $t('drama.createDesc') }}</div>
    
    <el-form 
      ref="formRef" 
      :model="form" 
      :rules="rules" 
      label-position="top"
      class="create-form"
      @submit.prevent="handleSubmit"
    >
      <!-- 1. Project Name (Required) -->
      <el-form-item :label="$t('drama.projectName')" prop="title" required>
        <el-input 
          v-model="form.title" 
          :placeholder="$t('drama.projectNamePlaceholder')"
          size="large"
          maxlength="100"
          show-word-limit
        />
      </el-form-item>

      <!-- 2. Project Description (Optional) -->
      <el-form-item :label="$t('drama.projectDesc')" prop="description">
        <el-input 
          v-model="form.description" 
          type="textarea" 
          :rows="3"
          :placeholder="$t('drama.projectDescPlaceholder')"
          maxlength="500"
          show-word-limit
          resize="none"
        />
      </el-form-item>

      <!-- 3. Reference Work (Optional) -->
      <el-form-item label="风格参考作品 (Style Reference Work)" prop="reference_work">
        <el-input 
          v-model="form.reference_work" 
          placeholder="例如：《七龙珠》、《宫崎骏动画》等，将自动同步到全局提示词"
          size="large"
          maxlength="200"
        />
      </el-form-item>

      <!-- 4. Visual Style (Dropdown) -->
      <el-form-item label="画面风格 (Visual Style)" prop="style">
        <el-select 
          v-model="form.style" 
          placeholder="选择画面风格" 
          size="large"
          class="style-select"
          @change="handleStyleChange"
        >
          <el-option
            v-for="option in styleOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>
        
        <!-- Style Preview Image -->
        <div class="style-preview-area mt-4" v-if="selectedStyleOption">
          <div class="style-large-preview">
            <el-image 
              :src="selectedStyleOption.image" 
              fit="cover"
              loading="lazy"
              class="preview-image"
              :preview-src-list="[selectedStyleOption.image]"
              :preview-teleported="true"
              hide-on-click-modal
            >
              <template #placeholder>
                <div class="image-placeholder" :style="{ background: selectedStyleOption.color }">
                  <el-icon class="loading-icon"><Loading /></el-icon>
                </div>
              </template>
              <template #error>
                <div class="image-error" :style="{ background: selectedStyleOption.color }">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
            <div class="zoom-hint">
              <el-icon><ZoomIn /></el-icon> 点击放大
            </div>
          </div>
          <div class="style-info-large">
            <p class="style-description">{{ selectedStyleOption.desc }}</p>
          </div>
        </div>
      </el-form-item>

      <!-- 5. Aspect Ratio -->
      <el-form-item label="画面比例" prop="aspect_ratio" required>
        <div class="aspect-ratio-selector">
          <div 
            class="ratio-option" 
            :class="{ active: form.aspect_ratio === '16:9' }"
            @click="form.aspect_ratio = '16:9'"
          >
            <div class="ratio-preview landscape"></div>
            <span class="ratio-label">横屏漫剧 (16:9)</span>
          </div>
          <div 
            class="ratio-option" 
            :class="{ active: form.aspect_ratio === '9:16' }"
            @click="form.aspect_ratio = '9:16'"
          >
            <div class="ratio-preview portrait"></div>
            <span class="ratio-label">竖屏漫剧 (9:16)</span>
          </div>
        </div>
      </el-form-item>

    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleClose">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button 
          type="primary" 
          size="large"
          :loading="loading"
          @click="handleSubmit"
        >
          <el-icon v-if="!loading"><Plus /></el-icon>
          {{ $t('drama.createNew') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Picture, Loading, ZoomIn } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import type { CreateDramaRequest } from '@/types/drama'
import celShadedImg from '@/assets/images/styles/cel_shaded_anime.png'
import makotoShinkaiImg from '@/assets/images/styles/makoto_shinkai.png'
import ghibliImg from '@/assets/images/styles/ghibli_style.png'
import shanghaiInkImg from '@/assets/images/styles/shanghai_ink_animation.png'
import impastoFantasyImg from '@/assets/images/styles/impasto_fantasy.png'
import magicalGirlImg from '@/assets/images/styles/magical_girl.png'
import cyberpunkImg from '@/assets/images/styles/cyberpunk.png'
import chibiImg from '@/assets/images/styles/chibi.png'
import lightNovelImg from '@/assets/images/styles/light_novel.png'
import americanAnimationImg from '@/assets/images/styles/american_animation.png'
import chineseAnimationImg from '@/assets/images/styles/chinese_animation.png'

/**
 * CreateDramaDialog - Reusable dialog for creating new drama projects
 * 创建短剧弹窗 - 可复用的创建短剧项目弹窗
 */
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'created': [id: string]
}>()

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

// v-model binding / 双向绑定
const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

// Style options / 风格选项
const styleOptions = [
  {
    label: '赛璐璐日漫',
    value: 'Cel-shaded Anime',
    image: celShadedImg,
    color: 'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
    desc: '经典赛璐璐风格，线条清晰，色块分明',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, official anime artwork, key visual, cel-shaded, classic cel animation technique, flat color blocks, crisp and clean line art, bold black outlines, hard-edged shadows, no gradient fills, vintage Japanese anime aesthetic, retro anime style, 2D hand-drawn animation style, soft ambient lighting, vibrant yet cohesive colors, professional composition, clean background, distinct color layers',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, bad hands, missing fingers, extra digits, fewer digits, text, signature, watermark, username, artist name, stamp, 3D, realistic, photorealistic, cgi, render, photography, messy line art, rough sketch, excessive color gradients, muddled colors, oversaturated colors, faded colors, multiple views'
  },
  {
    label: '新海诚电影风格',
    value: 'Makoto Shinkai Style',
    image: makotoShinkaiImg,
    color: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
    desc: '唯美光影，细腻写实，恢弘大气，每一帧都是壁纸',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, official anime movie key visual, Makoto Shinkai animation style, cinematic shot, dynamic camera angle, breathtaking sky, towering cumulus clouds, golden hour, twilight, aurora, meteor shower, god rays, crepuscular rays, volumetric light, atmospheric perspective, subtle lens flare, bokeh, light refraction, water reflection, raindrops on window, dewdrops, iridescent luster, soft pastel colors, harmonious color grading, ethereal atmosphere, nostalgic emotion, hyper-detailed environments, realistic architecture, lush natural scenery, delicate character rendering, soft edge shadows, film grain, subtle chromatic aberration, movie-like depth of field, 2D hand-drawn anime, seamless color blending',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, bad hands, missing fingers, extra digits, fewer digits, text, signature, watermark, username, artist name, stamp, logo, 3D, realistic, photorealistic, cgi, render, photography, messy line art, rough sketch, flat colors, dull tones, harsh lighting, no depth, overexposed, underexposed, multiple views, inconsistent details'
  },
  {
    label: '吉卜力/宫崎骏风格',
    value: 'Studio Ghibli Style',
    image: ghibliImg,
    color: 'linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%)',
    desc: '温暖治愈，手绘质感，自然清新，童话般的梦幻氛围',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, Studio Ghibli key visual, Hayao Miyazaki signature style, traditional hand-painted background, 2D hand-drawn cel animation, soft warm color grading, gentle sunlight through foliage, dappled light, delicate texture rendering, rich natural elements, moss, vines, wildflowers, rustic architecture, fluffy clouds, nostalgic atmosphere, magical ethereal feeling, subtle film grain, no harsh lines, smooth outlines, cozy and warm ambiance, depth of field',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, missing fingers, extra digits, text, signature, watermark, logo, stamp, 3D, realistic, photorealistic, cgi, render, photography, messy lines, rough sketch, harsh shadows, sharp edges, excessive gradients, muddled colors, overexposed, dull hues, inorganic designs, multiple views'
  },
  {
    label: '欧美动画风格',
    value: 'American Animation Style',
    image: americanAnimationImg,
    color: 'linear-gradient(135deg, #3a7bd5 0%, #3a6073 100%)',
    desc: '迪士尼/皮克斯风格，极致细节，生动表情，好莱坞大片质感',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, blockbuster 3D American animated film key visual, Pixar-level CGI, stylized realistic 3D rendering, hyper-detailed material textures (fur, fabric, skin, metal), subsurface scattering, cinematic global illumination, beautiful rim lighting, soft depth of field, bokeh, expressive micro-expressions, polished 3D model, dynamic and natural poses, atmospheric haze, rich and harmonious color grading, magical and lively atmosphere, theatrical animation quality, professional composition',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, poorly rendered hands, missing fingers, extra digits, text, signature, watermark, logo, stamp, 2D cartoon, japanese anime, photorealistic photography, low-poly models, flat colors, harsh lighting, multiple views, inconsistent textures'
  },
  {
    label: '国漫风格',
    value: 'Chinese Animation Style',
    image: chineseAnimationImg,
    color: 'linear-gradient(135deg, #ff9966 0%, #ff5e62 100%)',
    desc: '顶级国漫视觉，精致唯美，东方韵味，大片质感',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, top-tier Chinese animation key visual, premium guoman art style, hyper-detailed oriental facial features, flawless skin rendering, intricate oriental costume and accessories, elegant and dynamic pose, cinematic global illumination, beautiful rim lighting, soft depth of field, bokeh, delicate fabric and texture rendering, oriental traditional patterns, misty clouds, distant mountains, ancient architectural elements, rich color grading, immersive oriental atmosphere, professional gallery quality, commercial print ready, smooth refined lines, soft layered shading',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, poorly drawn hands, missing fingers, extra digits, text, signature, watermark, logo, stamp, realistic photography, Western animation elements, Japanese anime style, low-poly models, flat shading, dull muted colors, messy background, distorted proportions, multiple views, inconsistent textures'
  },
  {
    label: '上美水墨动画风',
    value: 'Shanghai Ink Animation',
    image: shanghaiInkImg,
    color: 'linear-gradient(135deg, #cfd9df 0%, #e2ebf0 100%)',
    desc: '中国传统水墨，虚实相生，气韵生动，独特的东方美学',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, Shanghai Animation Film Studio classic ink wash animation, authentic Chinese ink wash painting style, traditional hand-drawn oriental animation, meticulous freehand brushwork, natural ink diffusion and halo, layered ink tones, Chinese blank space aesthetic, silk and rice paper texture, retro animation film grain, soft ink transitions, no rigid outlines, elegant ink and light color matching, distant mountains, misty clouds, flowing water texture, timeless oriental artistic conception, peaceful and remote atmosphere, harmonious traditional Chinese composition',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, missing fingers, extra digits, text, signature, watermark, stamp, 3D, realistic, photorealistic, cgi, render, photography, sharp edges, hard line art, cel shading, solid color blocks, oversaturated colors, neon colors, modern digital painting style, Western anime elements, messy ink stains, inconsistent ink layers, multiple views'
  },
  {
    label: '厚涂幻想风',
    value: 'Impasto Fantasy Style',
    image: impastoFantasyImg,
    color: 'linear-gradient(135deg, #4b6cb7 0%, #182848 100%)',
    desc: '史诗质感，厚重笔触，光影戏剧，魔幻艺术的极致表现',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, AAA game key visual, fantasy concept art, classic impasto oil painting technique, heavy textured brushstrokes, layered paint application, visible painterly strokes, dramatic cinematic lighting, god rays, crepuscular rays, hard and soft shadows, subsurface scattering, hyper-realistic material details (metal, leather, fur, stone, magical energy), iridescent magical glows, moody atmospheric perspective, rich contrasting colors, harmonious color grading, epic heroic atmosphere, mystical fantasy vibe, intricate background details, solid anatomical volume, artistic imperfection brush texture, professional gallery quality',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, bad hands, missing fingers, extra digits, text, signature, watermark, stamp, flat art, cel shading, vector art, clean line art, smooth digital painting, no brush texture, 3D, photorealistic, cgi, render, photography, harsh neon colors, washed-out colors, flat lighting, no depth, multiple views, low-poly, cartoonish simplification'
  },
  {
    label: '魔法少女风格',
    value: 'Magical Girl Style',
    image: magicalGirlImg,
    color: 'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
    desc: '华丽变身，梦幻魔法，经典少女漫风格',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, iconic 90s shoujo anime, magical girl key visual, Sailor Moon official art style, hand-drawn 2D anime, hyper-detailed magical costume, intricate lace and tulle details, crystal and pearl embellishments, golden trimmings, flowing satin ribbons, sparkling glitter effects, radiant star bursts, glowing crescent moon, rainbow refraction magic, magical wand, dynamic action pose, mid-transformation scene, volumetric magical light, soft backlighting, vibrant gradient colors, harmonious color grading, dreamy starry sky, floating petals, subtle film grain, crisp clean outlines, heroic and ethereal atmosphere, professional composition',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, bad hands, missing fingers, extra digits, text, signature, watermark, logo, stamp, 3D, realistic, photorealistic, cgi, render, photography, messy lines, harsh shadows, flat colors, dull tones, no magical sparkles, plain clothing, multiple views, modern minimalist style'
  },
  {
    label: '科幻赛博朋克风格',
    value: 'Sci-Fi Cyberpunk Style',
    image: cyberpunkImg,
    color: 'linear-gradient(135deg, #141E30 0%, #243B55 100%)',
    desc: '未来科技，霓虹闪烁，反乌托邦世界的视觉冲击',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, cyberpunk key visual, sci-fi blockbuster cinematic shot, dystopian future metropolis, glowing neon signage in japanese and english, holographic projections, floating digital billboards, massive megabuildings, flying vehicles and airships, heavy downpour, water puddles with light reflections, atmospheric fog and steam vents, dramatic rim lighting, god rays through skyscrapers, gritty textured surfaces, rusted metal, polished chrome, damaged concrete, cybernetic enhancements, glitch art effects, lens flare, bokeh, moody atmospheric perspective, dark noir tone, high detail textures, professional concept art, realistic material rendering, dystopian atmosphere',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, bad hands, missing fingers, extra digits, text, signature, watermark, username, stamp, cartoon, anime cel shading, flat colors, bright pastels, fantasy magic, medieval elements, clean minimalist design, multiple views, unrealistic lighting, oversaturated colors'
  },
  {
    label: 'Q版风格',
    value: 'Chibi Style',
    image: chibiImg,
    color: 'linear-gradient(135deg, #FF9A9E 0%, #FECFEF 100%)',
    desc: '可爱萌系，Q版造型，治愈人心',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, premium chibi illustration, kawaii chibi art, super deformed SD style, perfect 2-4 head proportion, oversized round head, petite cute body, huge expressive eyes, delicate cute facial features, subtle blush, rounded soft outlines, smooth clean line art, soft gradient shading, fluffy and smooth textures, vibrant but gentle color matching, cute decorative elements, soft ambient lighting, dreamy cute atmosphere, professional composition, no sharp edges',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, realistic adult proportions, long thin limbs, sharp corners, harsh hard lines, text, signature, watermark, stamp, 3D photorealism, cgi render, photography, dull muted colors, overly complex details, multiple views, inconsistent proportions'
  },
  {
    label: '轻小说封面插画风',
    value: 'Light Novel Cover Style',
    image: lightNovelImg,
    color: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)',
    desc: '精致唯美，极具张力的角色插画',
    prompt: 'masterpiece, best quality, 8K, ultra high resolution, extremely detailed, top-tier light novel cover illustration, japanese modern anime illustration style, central focused main character, striking character design, hyper-detailed facial features, flawless skin rendering, intricate costume and accessories, dynamic and elegant pose, cinematic key visual composition, golden hour lighting, volumetric light, beautiful rim lighting, lens flare, bokeh, soft ambient occlusion, rich and harmonious color palette, iridescent color accents, subtle depth of field, atmospheric haze, minimalist immersive background that complements character, dreamy and immersive atmosphere, professional gallery quality, commercial print ready, smooth refined line art, soft gradient shading, realistic fabric and texture rendering',
    negative_prompt: 'lowres, worst quality, low quality, normal quality, jpeg artifacts, pixelated, blurry, out of frame, cropped, ugly, deformed, disfigured, bad anatomy, poorly drawn hands, missing fingers, extra digits, text, signature, watermark, logo, stamp, 3D render, photorealism, cgi, photography, distorted proportions, messy cluttered composition, background overwhelming main character, flat colors, dull tones, harsh shadows, multiple distracting characters, low detail, sketchy lines, inconsistent textures, multiple views'
  }
]

// Form data / 表单数据
const form = reactive<CreateDramaRequest>({
  title: '',
  description: '',
  style: 'Cel-shaded Anime',
  reference_work: '',
  reference_image: '',
  style_prompt: styleOptions[0].prompt,
  aspect_ratio: '16:9'
})

const selectedStyleOption = computed(() => {
  return styleOptions.find(opt => opt.value === form.style) || styleOptions[0]
})

const handleStyleChange = (val: string) => {
  const option = styleOptions.find(opt => opt.value === val)
  if (option) {
    form.style_prompt = option.prompt
  }
}

// Validation rules / 验证规则
const rules: FormRules = {
  title: [
    { required: true, message: '请输入项目标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  aspect_ratio: [
    { required: true, message: '请选择画面比例', trigger: 'change' }
  ]
}

// Reset form when dialog closes / 关闭时重置表单
const handleClosed = () => {
  form.title = ''
  form.description = ''
  form.style = 'Cel-shaded Anime'
  form.reference_work = ''
  form.reference_image = ''
  form.style_prompt = styleOptions[0].prompt
  form.aspect_ratio = '16:9'
  formRef.value?.resetFields()
}

// Close dialog / 关闭弹窗
const handleClose = () => {
  visible.value = false
}

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    loading.value = true
    try {
      const drama = await dramaAPI.create(form)
      ElMessage.success('创建成功')
      visible.value = false
      emit('created', drama.id)
      // Navigate to drama detail page / 跳转到短剧详情页
      router.push(`/dramas/${drama.id}`)
    } catch (error: any) {
      ElMessage.error(error.message || '创建失败')
    } finally {
      loading.value = false
    }
  } catch (error) {
    console.warn('Validation failed', error)
  }
}
</script>

<style scoped>
/* ========================================
   Dialog Styles / 弹窗样式
   ======================================== */
.create-dialog :deep(.el-dialog) {
  border-radius: var(--radius-xl);
}

.create-dialog :deep(.el-dialog__header) {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.create-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.create-dialog :deep(.el-dialog__body) {
  padding: 1.5rem;
}

.create-dialog :deep(.el-dialog__footer) {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-primary);
}

.dialog-desc {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
  line-height: 1.5;
}

.create-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

/* ========================================
   Style Select / 风格选择
   ======================================== */
.style-select {
  width: 100%;
}

.style-hint {
  margin-top: 0.5rem;
  font-size: 0.75rem;
  color: var(--warning-color);
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.style-preview-area {
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  background-color: var(--bg-secondary);
}

.style-large-preview {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  overflow: hidden;
  cursor: zoom-in;
}

.preview-image {
  width: 100%;
  height: 100%;
  transition: transform 0.3s ease;
}

.style-large-preview:hover .preview-image {
  transform: scale(1.05);
}

.zoom-hint {
  position: absolute;
  bottom: 0.5rem;
  right: 0.5rem;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.style-large-preview:hover .zoom-hint {
  opacity: 1;
}

.style-info-large {
  padding: 1rem;
  background: white;
  border-top: 1px solid var(--border-primary);
}

.style-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
  margin: 0;
}

/* ========================================
   Aspect Ratio / 画面比例
   ======================================== */
.aspect-ratio-selector {
  display: flex;
  gap: 1rem;
}

.ratio-option {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--bg-secondary);
}

.ratio-option:hover {
  border-color: var(--primary-color-light);
  background: var(--primary-color-lighter);
}

.ratio-option.active {
  border-color: var(--primary-color);
  background: var(--primary-color-lighter);
  color: var(--primary-color);
}

.ratio-preview {
  background: #e2e8f0;
  border: 2px solid #cbd5e1;
  border-radius: 4px;
}

.ratio-preview.landscape {
  width: 64px;
  height: 36px;
}

.ratio-preview.portrait {
  width: 36px;
  height: 64px;
}

.ratio-label {
  font-size: 0.875rem;
  font-weight: 500;
}

/* ========================================
   Responsive / 响应式
   ======================================== */
@media (max-width: 640px) {
  .create-dialog :deep(.el-dialog) {
    width: 90% !important;
    margin-top: 5vh !important;
  }
}
</style>