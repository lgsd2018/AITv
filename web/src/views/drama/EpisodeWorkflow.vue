<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="$router.back()" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ $t('workflow.backToProject') }}</span>
          </el-button>
          <h1 class="header-title">{{ $t('workflow.episodeProduction', { number: episodeNumber }) }}</h1>
        </template>
        <template #center>
          <div class="custom-steps">
            <div class="step-item" :class="{ active: currentStep >= 0, current: currentStep === 0 }">
              <div class="step-circle">1</div>
              <span class="step-text">{{ $t('workflow.steps.content') }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div class="step-item" :class="{ active: currentStep >= 1, current: currentStep === 1 }">
              <div class="step-circle">2</div>
              <span class="step-text">{{ $t('workflow.steps.generateImages') }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div class="step-item" :class="{ active: currentStep >= 2, current: currentStep === 2 }">
              <div class="step-circle">3</div>
              <span class="step-text">{{ $t('workflow.steps.splitStoryboard') }}</span>
            </div>
          </div>
        </template>
        <template #right>
          <el-button :icon="Setting" @click="showModelConfigDialog" :title="$t('workflow.modelConfig')">
            图文配置
          </el-button>
        </template>
      </AppHeader>

    <div class="project-info-section animate-fade-in" v-if="drama" style="margin-bottom: 20px;">
      <el-descriptions :column="4" border size="small" class="info-descriptions">
        <el-descriptions-item label="项目名称">{{ drama.title }}</el-descriptions-item>
        <el-descriptions-item label="画面风格">{{ getStyleLabel(drama.style) }}</el-descriptions-item>
        <el-descriptions-item label="参考作品">{{ drama.reference_work || '无' }}</el-descriptions-item>
        <el-descriptions-item label="画面比例">{{ drama.aspect_ratio || '16:9' }}</el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- 阶段 0: 章节内容 + 提取角色场景 -->
    <el-card v-show="currentStep === 0" shadow="never" class="stage-card stage-card-fullscreen">
      <div class="stage-body stage-body-fullscreen">
        <!-- 未保存时显示输入框 -->
        <div v-if="!hasScript" class="generation-form">
          <el-input
            v-model="scriptContent"
            type="textarea"
:placeholder="$t('workflow.scriptPlaceholder')"
            class="script-textarea script-textarea-fullscreen"
          />

          <div class="action-buttons-inline">
            <el-button 
              type="primary" 
              size="default" 
              @click="saveChapterScript"
              :disabled="!scriptContent.trim() || generatingScript"
            >
              <el-icon><Check /></el-icon>
              <span>{{ $t('workflow.saveChapter') }}</span>
            </el-button>
          </div>
        </div>

        <!-- 已保存时显示内容 -->
        <div v-if="hasScript" class="overview-section">
          <div class="episode-info">
            <h3>{{ $t('workflow.chapterContent', { number: episodeNumber }) }}</h3>
            <el-tag type="success" size="large">{{ $t('workflow.saved') }}</el-tag>
          </div>
          <div class="overview-content">
            <el-input 
              v-model="currentEpisode.script_content"
              type="textarea"
              :rows="15"
              readonly
              class="script-display"
            />
          </div>

          <el-divider />

          <!-- 显示已提取的角色和场景 -->
          <div v-if="hasExtractedData" class="extracted-info">
            <el-alert 
              type="success" 
              :closable="false"
              style="margin-bottom: 16px;"
            >
              <template #title>
                <div style="display: flex; align-items: center; gap: 16px;">
                  <span>✅ {{ $t('workflow.extractedData') }}</span>
                  <el-tag v-if="hasCharacters" type="success">{{ $t('workflow.characters') }}: {{ charactersCount }}</el-tag>
                  <el-tag v-if="currentEpisode?.scenes" type="success">{{ $t('workflow.scenes') }}: {{ currentEpisode.scenes.length }}</el-tag>
                </div>
              </template>
            </el-alert>
            
            <!-- 角色列表 -->
            <div v-if="hasCharacters" style="margin-bottom: 16px;">
              <h4 class="extracted-title">{{ $t('workflow.extractedCharacters') }}：</h4>
              <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                <el-tag 
                  v-for="char in currentEpisode?.characters" 
                  :key="char.id"
                  type="info"
                >
                  {{ char.name }} <span v-if="char.role" class="secondary-text">({{ char.role }})</span>
                </el-tag>
              </div>
            </div>
            
            <!-- 场景列表 -->
            <div v-if="currentEpisode?.scenes && currentEpisode.scenes.length > 0">
              <h4 class="extracted-title">{{ $t('workflow.extractedScenes') }}：</h4>
              <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                <el-tag 
                  v-for="scene in currentEpisode.scenes" 
                  :key="scene.id"
                  type="warning"
                >
                  {{ scene.location }} <span class="secondary-text">· {{ scene.time }}</span>
                </el-tag>
              </div>
            </div>
          </div>

          <el-divider />

          <div class="action-buttons">
            <el-button 
              type="primary"
              size="large"
              @click="handleExtractCharactersAndBackgrounds"
              :loading="extractingCharactersAndBackgrounds"
              :disabled="!hasScript"
            >
              <el-icon><MagicStick /></el-icon>
              {{ hasExtractedData ? $t('workflow.reExtract') : $t('workflow.extractCharactersAndScenes') }}
            </el-button>
            <el-button 
              type="success"
              size="large"
              @click="nextStep"
              :disabled="!hasExtractedData"
            >
              {{ $t('workflow.nextStepGenerateImages') }}
              <el-icon><ArrowRight /></el-icon>
            </el-button>
            <div v-if="!hasExtractedData" style="margin-top: 8px;">
              <el-alert type="warning" :closable="false" style="display: inline-block;">
                <template #title>
                  <span style="font-size: 12px;">
                    {{ $t('workflow.extractWarning') }}
                  </span>
                </template>
              </el-alert>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 阶段 1: 生成图片 -->
    <el-card v-show="currentStep === 1" class="workflow-card">
      <div class="stage-body">
        <!-- 角色图片生成 -->
        <div class="image-gen-section">
          <div class="section-header">
            <div class="section-title">
              <h3>
                <el-icon><User /></el-icon>
                {{ $t('workflow.characterImages') }}
              </h3>
              <el-alert 
                type="info"
                :closable="false"
                style="margin: 0;"
              >
                {{ $t('workflow.characterCount', { count: charactersCount }) }}
              </el-alert>
            </div>
            <div class="section-actions">
              <el-checkbox 
                v-model="selectAllCharacters"
                @change="toggleSelectAllCharacters"
                style="margin-right: 12px;"
              >
                {{ $t('workflow.selectAll') }}
              </el-checkbox>
              <el-button 
                type="primary"
                @click="batchGenerateCharacterImages"
                :loading="batchGeneratingCharacters"
                :disabled="selectedCharacterIds.length === 0"
                size="default"
              >
                {{ $t('workflow.batchGenerate') }} ({{ selectedCharacterIds.length }})
              </el-button>
            </div>
          </div>
          
          <div class="character-image-list">
            <div v-for="char in currentEpisode?.characters" :key="char.id" class="character-item">
              <el-card shadow="hover" class="fixed-card">
                <div class="card-header">
                  <el-checkbox 
                    v-model="selectedCharacterIds"
                    :value="char.id"
                    style="margin-right: 8px;"
                  />
                  <div class="header-left">
                    <h4>{{ char.name }}</h4>
                    <el-tag size="small">{{ char.role }}</el-tag>
                  </div>
                  <el-button 
                    type="danger" 
                    size="small" 
                    :icon="Delete"
                    circle
                    @click="deleteCharacter(char.id)"
:title="$t('workflow.deleteCharacter')"
                  />
                </div>
                
                <div class="card-image-container">
                  <div v-if="char.image_url" class="char-image">
                    <el-image 
                      :src="char.image_url" 
                      fit="cover" 
                      :preview-src-list="[char.image_url]"
                      :preview-teleported="true"
                    />
                  </div>
                  <div v-else-if="char.image_generation_status === 'pending' || char.image_generation_status === 'processing' || generatingCharacterImages[char.id]" class="char-placeholder generating">
                    <el-icon :size="64" class="rotating"><Loading /></el-icon>
                    <span>{{ $t('common.generating') }}</span>
                    <el-tag type="warning" size="small" style="margin-top: 8px;">{{ char.image_generation_status === 'pending' ? $t('common.queuing') : $t('common.processing') }}</el-tag>
                  </div>
                  <div v-else-if="char.image_generation_status === 'failed'" class="char-placeholder failed">
                    <el-icon :size="64"><WarningFilled /></el-icon>
                    <span>{{ $t('common.generateFailed') }}</span>
                    <el-tag type="danger" size="small" style="margin-top: 8px;">{{ $t('common.clickToRegenerate') }}</el-tag>
                  </div>
                  <div v-else class="char-placeholder">
                    <el-icon :size="64"><User /></el-icon>
                    <span>{{ $t('common.notGenerated') }}</span>
                  </div>
                </div>

                <div class="card-actions">
                  <el-tooltip :content="$t('tooltip.editPrompt')" placement="top">
                    <el-button 
                      size="small" 
                      @click="openPromptDialog(char, 'character')"
                      :icon="Edit"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.aiGenerate')" placement="top">
                    <el-button 
                      type="primary"
                      size="small" 
                      @click="generateCharacterImage(char.id)"
                      :loading="generatingCharacterImages[char.id]"
                      :icon="MagicStick"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.uploadImage')" placement="top">
                    <el-button 
                      size="small" 
                      @click="uploadCharacterImage(char.id)"
                      :icon="Upload"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.selectFromLibrary')" placement="top">
                    <el-button 
                      size="small" 
                      @click="selectFromLibrary(char.id)"
                      :icon="Picture"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('workflow.addToLibrary')" placement="top">
                    <el-button 
                      size="small" 
                      @click="addToCharacterLibrary(char)"
                      :icon="FolderAdd"
                      :disabled="!char.image_url"
                      circle
                    />
                  </el-tooltip>
                </div>
              </el-card>
            </div>
          </div>
        </div>

        <el-divider />

        <!-- 场景图片生成 -->
        <div class="image-gen-section">
          <div class="section-header">
            <div class="section-title">
              <h3>
                <el-icon><Place /></el-icon>
                {{ $t('workflow.sceneImages') }}
              </h3>
              <el-alert 
                type="info"
                :closable="false"
                style="margin: 0;"
              >
                {{ $t('workflow.sceneCount', { count: drama?.scenes?.length || 0 }) }}
              </el-alert>
            </div>
            <div class="section-actions">
              <el-checkbox 
                v-model="selectAllScenes"
                @change="toggleSelectAllScenes"
                style="margin-right: 12px;"
              >
                {{ $t('workflow.selectAll') }}
              </el-checkbox>
              <el-button 
                type="primary"
                @click="batchGenerateSceneImages"
                :loading="batchGeneratingScenes"
                :disabled="selectedSceneIds.length === 0"
                size="default"
              >
                {{ $t('workflow.batchGenerateSelected') }} ({{ selectedSceneIds.length }})
              </el-button>
            </div>
          </div>
          
          <div class="scene-image-list">
            <div v-for="scene in currentEpisode?.scenes" :key="scene.id" class="scene-item">
              <el-card shadow="hover" class="fixed-card">
                <div class="card-header">
                  <el-checkbox 
                    v-model="selectedSceneIds"
                    :value="scene.id"
                    style="margin-right: 8px;"
                  />
                  <div class="header-left">
                    <h4>{{ scene.location }}</h4>
                    <el-tag size="small">{{ scene.time }}</el-tag>
                  </div>
                </div>

                <div class="card-image-container">
                  <div v-if="scene.image_url" class="scene-image">
                    <el-image 
                      :src="scene.image_url" 
                      fit="cover" 
                      :preview-src-list="[scene.image_url]"
                      :preview-teleported="true"
                    />
                  </div>
                  <div v-else-if="scene.image_generation_status === 'pending' || scene.image_generation_status === 'processing' || generatingSceneImages[scene.id]" class="scene-placeholder generating">
                    <el-icon :size="64" class="rotating"><Loading /></el-icon>
                    <span>{{ $t('common.generating') }}</span>
                    <el-tag type="warning" size="small" style="margin-top: 8px;">{{ scene.image_generation_status === 'pending' ? $t('common.queuing') : $t('common.processing') }}</el-tag>
                  </div>
                  <div v-else-if="scene.image_generation_status === 'failed'" class="scene-placeholder failed" @click="generateSceneImage(scene.id)" style="cursor: pointer;">
                    <el-icon :size="64"><WarningFilled /></el-icon>
                    <span>{{ $t('common.generateFailed') }}</span>
                    <el-tag type="danger" size="small" style="margin-top: 8px;">{{ $t('common.clickToRegenerate') }}</el-tag>
                  </div>
                  <div v-else class="scene-placeholder">
                    <el-icon :size="64"><Place /></el-icon>
                    <span>{{ $t('common.notGenerated') }}</span>
                  </div>
                </div>

                <div class="card-actions">
                  <el-tooltip :content="$t('tooltip.editPrompt')" placement="top">
                    <el-button 
                      size="small" 
                      @click="openPromptDialog(scene, 'scene')"
                      :icon="Edit"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.aiGenerate')" placement="top">
                    <el-button 
                      type="primary"
                      size="small" 
                      @click="generateSceneImage(scene.id)"
                      :loading="generatingSceneImages[scene.id]"
                      :icon="MagicStick"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.uploadImage')" placement="top">
                    <el-button 
                      size="small" 
                      @click="uploadSceneImage(scene.id)"
                      :icon="Upload"
                      circle
                    />
                  </el-tooltip>
                </div>
              </el-card>
            </div>
          </div>
        </div>

        <el-divider />

        <div class="action-buttons">
          <el-button size="large" @click="prevStep">
            <el-icon><ArrowLeft /></el-icon>
            {{ $t('workflow.prevStep') }}
          </el-button>
          <el-button 
            type="success"
            size="large"
            @click="nextStep"
            :disabled="!allImagesGenerated"
          >
            {{ $t('workflow.nextStepSplitShots') }}
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          <div v-if="!allImagesGenerated" style="margin-top: 8px;">
            <el-alert type="warning" :closable="false" style="display: inline-block;">
              <template #title>
                <span style="font-size: 12px;">
                  {{ $t('workflow.generateAllImagesFirst') }}
                </span>
              </template>
            </el-alert>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 阶段 2: 拆分分镜 -->
    <el-card v-show="currentStep === 2" shadow="never" class="stage-card">
      <div class="stage-body">
        <!-- 分镜列表 -->
        <div v-if="currentEpisode?.storyboards && currentEpisode.storyboards.length > 0" class="shots-list">
          <div class="shots-header">
            <h3>{{ $t('workflow.shotList') }}</h3>
          </div>
          
          <el-table :data="currentEpisode.storyboards" border stripe style="margin-top: 16px;">
            <el-table-column type="index" :label="$t('storyboard.table.number')" width="60" />
            <el-table-column :label="$t('storyboard.table.title')" width="120" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.title || '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.shotType')" width="80">
              <template #default="{ row }">
                {{ row.shot_type || '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.movement')" width="80">
              <template #default="{ row }">
                {{ row.movement || '-' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.location')" width="150">
              <template #default="{ row }">
                <el-popover 
                  placement="right" 
                  :width="300" 
                  trigger="hover"
                  :content="row.action || '-'"
                >
                  <template #reference>
                    <!-- 单行打点 -->
                    <span class="overflow-tooltip">{{ row.location || '-' }}</span>
                  </template>
                </el-popover>
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.character')" width="100">
              <template #default="{ row }">
                <span v-if="row.characters && row.characters.length > 0">
                  {{ row.characters.map(c => c.name || c).join(', ') }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.action')">
              <template #default="{ row }">
                <el-popover 
                  placement="right" 
                  :width="300" 
                  trigger="hover"
                  :content="row.action || '-'"
                >
                  <template #reference>
                    <!-- 单行打点 -->
                    <span class="overflow-tooltip">{{ row.action || '-' }}</span>
                  </template>
                </el-popover>
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.duration')" width="80">
              <template #default="{ row }">
                {{ row.duration || '-' }}秒
              </template>
            </el-table-column>
            <el-table-column :label="$t('storyboard.table.operations')" width="100" fixed="right">
              <template #default="{ row, $index }">
                <el-button 
                  type="primary" 
                  size="small"
                  @click="editShot(row, $index)"
                >
                  {{ $t('common.edit') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div class="action-buttons" style="margin-top: 24px;">
            <el-button size="large" @click="prevStep">
              <el-icon><ArrowLeft /></el-icon>
              {{ $t('workflow.prevStep') }}
            </el-button>
            <el-button 
              @click="regenerateShots"
              :icon="MagicStick"
            >
              {{ $t('workflow.reSplitShots') }}
            </el-button>
            <el-button 
              type="success"
              size="large"
              @click="goToProfessionalUI"
            >
              {{ $t('workflow.enterProfessional') }}
              <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
        </div>
        
        <!-- 未拆分时显示 -->
        <div v-else class="empty-shots">
          <el-empty :description="$t('workflow.splitStoryboardFirst')">
            <el-button 
              type="primary" 
              @click="generateShots"
              :loading="generatingShots"
              :icon="MagicStick"
            >
              {{ generatingShots ? $t('workflow.aiSplitting') : $t('workflow.aiAutoSplit') }}
            </el-button>
            
            <!-- 任务进度显示 -->
            <div v-if="generatingShots" style="margin-top: 24px; max-width: 400px; margin-left: auto; margin-right: auto;">
              <el-progress :percentage="taskProgress" :status="taskProgress === 100 ? 'success' : undefined">
                <template #default="{ percentage }">
                  <span style="font-size: 12px;">{{ percentage }}%</span>
                </template>
              </el-progress>
              <div class="task-message">
                {{ taskMessage }}
              </div>
            </div>
          </el-empty>
        </div>
      </div>
    </el-card>

    <!-- 阶段 3: 专业制作（占位，实际跳转到专业UI页面） -->

    <!-- 镜头编辑对话框 -->
    <el-dialog 
      v-model="shotEditDialogVisible" 
:title="$t('workflow.editShot')" 
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form v-if="editingShot" label-width="100px" size="default">
        <el-form-item :label="$t('workflow.shotTitle')">
          <el-input v-model="editingShot.title" :placeholder="$t('workflow.shotTitlePlaceholder')" />
        </el-form-item>
        
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item :label="$t('workflow.shotType')">
              <el-select v-model="editingShot.shot_type" :placeholder="$t('workflow.selectShotType')">
                <el-option :label="$t('workflow.longShot')" value="远景" />
                <el-option :label="$t('workflow.fullShot')" value="全景" />
                <el-option :label="$t('workflow.mediumShot')" value="中景" />
                <el-option :label="$t('workflow.closeUp')" value="近景" />
                <el-option :label="$t('workflow.extremeCloseUp')" value="特写" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="$t('workflow.cameraAngle')">
              <el-select v-model="editingShot.angle" :placeholder="$t('workflow.selectAngle')">
                <el-option :label="$t('workflow.eyeLevel')" value="平视" />
                <el-option :label="$t('workflow.lowAngle')" value="仰视" />
                <el-option :label="$t('workflow.highAngle')" value="俯视" />
                <el-option :label="$t('workflow.sideView')" value="侧面" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="$t('workflow.cameraMovement')">
              <el-select v-model="editingShot.movement" :placeholder="$t('workflow.selectMovement')">
                <el-option :label="$t('workflow.staticShot')" value="固定镜头" />
                <el-option :label="$t('workflow.pushIn')" value="推镜" />
                <el-option :label="$t('workflow.pullOut')" value="拉镜" />
                <el-option :label="$t('workflow.followShot')" value="跟镜" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="$t('workflow.location')">
              <el-input v-model="editingShot.location" :placeholder="$t('workflow.locationPlaceholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('workflow.time')">
              <el-input v-model="editingShot.time" :placeholder="$t('workflow.timeSetting')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="$t('workflow.shotDescription')">
          <el-input v-model="editingShot.description" type="textarea" :rows="2" :placeholder="$t('workflow.shotDescriptionPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.actionDescription')">
          <el-input v-model="editingShot.action" type="textarea" :rows="3" :placeholder="$t('workflow.detailedAction')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.dialogue')">
          <el-input v-model="editingShot.dialogue" type="textarea" :rows="2" :placeholder="$t('workflow.characterDialogue')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.result')">
          <el-input v-model="editingShot.result" type="textarea" :rows="2" :placeholder="$t('workflow.actionResult')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.atmosphere')">
          <el-input v-model="editingShot.atmosphere" type="textarea" :rows="2" :placeholder="$t('workflow.atmosphereDescription')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.imagePrompt')">
          <el-input v-model="editingShot.image_prompt" type="textarea" :rows="3" :placeholder="$t('workflow.imagePromptPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('workflow.videoPrompt')">
          <el-input v-model="editingShot.video_prompt" type="textarea" :rows="3" :placeholder="$t('workflow.videoPromptPlaceholder')" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="$t('workflow.bgmHint')">
              <el-input v-model="editingShot.bgm_prompt" :placeholder="$t('workflow.bgmAtmosphere')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('workflow.soundEffect')">
              <el-input v-model="editingShot.sound_effect" :placeholder="$t('workflow.soundEffectDescription')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="$t('workflow.durationSeconds')">
          <el-input-number v-model="editingShot.duration" :min="1" :max="60" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="shotEditDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveShotEdit" :loading="savingShot">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- 提示词编辑对话框 -->
    <el-dialog 
      v-model="promptDialogVisible" 
:title="$t('workflow.editPrompt')" 
      width="600px"
    >
      <el-form label-width="80px">
        <el-form-item :label="$t('common.name')">
          <el-input v-model="currentEditItem.name" disabled />
        </el-form-item>

        <el-form-item label="画面风格">
          <el-input v-model="editStyle" disabled placeholder="继承自项目设置">
            <template #append>项目预设</template>
          </el-input>
        </el-form-item>

        <el-form-item label="参考作品">
          <el-input v-model="editReference" disabled placeholder="继承自项目设置">
            <template #append>项目预设</template>
          </el-input>
        </el-form-item>

        <el-form-item label="图片尺寸">
          <el-select v-model="editSize" style="width: 100%" disabled>
            <el-option label="16:9 (2560x1440)" value="2560x1440" />
            <el-option label="9:16 (1440x2560)" value="1440x2560" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="currentEditType === 'character'" label="辅助设置">
          <el-checkbox-group v-model="editViewSettings">
            <el-checkbox label="white_background">白底图 (White Background)</el-checkbox>
            <el-checkbox label="three_views">三视图 (Front, Side, Back)</el-checkbox>
            <el-checkbox label="composition_fix">构图修正 (Centered, No Crop)</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item :label="$t('workflow.imagePrompt')">
          <el-input
            v-model="editPrompt"
            type="textarea"
            :rows="6"
            :placeholder="$t('workflow.imagePromptPlaceholder')"
          />
          <!-- 一键优化与撤销按钮 -->
          <div style="margin-top: 8px; display: flex; gap: 8px; justify-content: flex-end;">
            <el-button size="small" @click="optimizePrompt">一键优化提示词</el-button>
            <el-button size="small" :disabled="!canUndoOptimize" @click="undoOptimizePrompt">撤销优化</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="savePrompt">{{ $t('common.saveAndGenerate') }}</el-button>
      </template>
    </el-dialog>

    <!-- 角色库选择对话框 -->
    <el-dialog 
      v-model="libraryDialogVisible" 
:title="$t('workflow.selectFromLibrary')" 
      width="800px"
    >
      <div class="library-grid">
        <div 
          v-for="item in libraryItems" 
          :key="item.id" 
          class="library-item"
          @click="selectLibraryItem(item)"
        >
          <el-image :src="item.image_url" fit="cover" />
          <div class="library-item-name">{{ item.name }}</div>
        </div>
      </div>
      <div v-if="libraryItems.length === 0" class="empty-library">
        <el-empty :description="$t('workflow.emptyLibrary')" />
      </div>
    </el-dialog>

    <!-- AI模型配置对话框 -->
    <el-dialog 
      v-model="modelConfigDialogVisible" 
:title="$t('workflow.aiModelConfig')" 
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form label-width="120px">
        <el-form-item :label="$t('workflow.textGenModel')">
          <el-select v-model="selectedTextModel" :placeholder="$t('workflow.selectTextModel')" style="width: 100%">
            <el-option 
              v-for="model in textModels" 
              :key="model.modelName" 
              :label="model.modelName"
              :value="model.modelName"
            />
          </el-select>
          <div class="model-tip">
            {{ $t('workflow.textModelTip') }}
          </div>
        </el-form-item>

        <el-form-item :label="$t('workflow.imageGenModel')">
          <el-select v-model="selectedImageModel" :placeholder="$t('workflow.selectImageModel')" style="width: 100%">
            <el-option 
              v-for="model in imageModels" 
              :key="model.modelName" 
              :label="model.modelName"
              :value="model.modelName"
            />
          </el-select>
          <div class="model-tip">
            {{ $t('workflow.modelConfigTip') }}
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="modelConfigDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveModelConfig">{{ $t('common.saveConfig') }}</el-button>
      </template>
    </el-dialog>

    <!-- 图片上传对话框 -->
    <el-dialog 
      v-model="uploadDialogVisible" 
:title="$t('tooltip.uploadImage')" 
      width="500px"
    >
      <el-upload
        class="upload-area"
        drag
        :action="uploadAction"
        :headers="uploadHeaders"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :show-file-list="false"
        accept="image/jpeg,image/png,image/jpg"
      >
        <el-icon class="el-icon--upload"><Upload /></el-icon>
        <div class="el-upload__text">
          {{ $t('workflow.dragFilesHere') }}<em>{{ $t('workflow.clickToUpload') }}</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            {{ $t('workflow.uploadFormatTip') }}
          </div>
        </template>
      </el-upload>
    </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  User, 
  Location, 
  Picture,
  MagicStick,
  ArrowRight,
  ArrowLeft,
  Place,
  Film,
  Edit,
  More,
  Upload,
  Delete,
  FolderAdd,
  Setting,
  Loading,
  WarningFilled
} from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { generationAPI } from '@/api/generation'
import { characterLibraryAPI } from '@/api/character-library'
import { aiAPI } from '@/api/ai'
import type { AIServiceConfig } from '@/types/ai'
import { imageAPI } from '@/api/image'
import type { Drama } from '@/types/drama'
import { AppHeader } from '@/components/common'

const route = useRoute()
const router = useRouter()
const { t: $t } = useI18n()
const dramaId = route.params.id as string
const episodeNumber = parseInt(route.params.episodeNumber as string)

const drama = ref<Drama>()

// Style label mapping
const styleLabelMap: Record<string, string> = {
  'Cel-shaded Anime': '赛璐璐日漫',
  'Makoto Shinkai Style': '新海诚电影风格',
  'Studio Ghibli Style': '吉卜力/宫崎骏风格',
  'American Animation Style': '欧美动画风格',
  'Chinese Animation Style': '国漫风格',
  'Shanghai Ink Animation': '上美水墨动画风',
  'Impasto Fantasy Style': '厚涂幻想风',
  'Magical Girl Style': '魔法少女风格',
  'Sci-Fi Cyberpunk Style': '科幻赛博朋克风格',
  'Chibi Style': 'Q版风格',
  'Light Novel Cover Style': '轻小说封面插画风',
  'realistic': '写实风格 (Realistic)'
}

const getStyleLabel = (style: string | undefined) => {
  if (!style) return '默认 (Realistic)'
  return styleLabelMap[style] || style
}

// 生成 localStorage key
const getStepStorageKey = () => `episode_workflow_step_${dramaId}_${episodeNumber}`

// 从 localStorage 恢复步骤，如果没有则默认为 0
const savedStep = localStorage.getItem(getStepStorageKey())
const currentStep = ref(savedStep ? parseInt(savedStep) : 0)
const scriptContent = ref('')
const generatingScript = ref(false)
const generatingShots = ref(false)
const extractingCharactersAndBackgrounds = ref(false)
const batchGeneratingCharacters = ref(false)
const batchGeneratingScenes = ref(false)
const generatingCharacterImages = ref<Record<number, boolean>>({})
const generatingSceneImages = ref<Record<number, boolean>>({})

// 选择状态
const selectedCharacterIds = ref<number[]>([])
const selectedSceneIds = ref<number[]>([])
const selectAllCharacters = ref(false)
const selectAllScenes = ref(false)

// 对话框状态
const promptDialogVisible = ref(false)
const libraryDialogVisible = ref(false)
const uploadDialogVisible = ref(false)
const modelConfigDialogVisible = ref(false)
const currentEditItem = ref<any>({ name: '' })
const currentEditType = ref<'character' | 'scene'>('character')
const editPrompt = ref('')
const editStyle = ref('')
const editReference = ref('')
const editSize = ref('')
const editViewSettings = ref<string[]>([])
// 一键优化提示词的撤销缓存
const optimizePromptHistory = ref('')
// 控制撤销按钮是否可用
const canUndoOptimize = ref(false)
const libraryItems = ref<any[]>([])
const currentUploadTarget = ref<any>(null)
const uploadAction = computed(() => '/api/v1/upload/image')
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
}))

// AI模型配置
interface ModelOption {
  modelName: string
  configName: string
  configId: number
  priority: number
}

const textModels = ref<ModelOption[]>([])
const imageModels = ref<ModelOption[]>([])
const selectedTextModel = ref<string>('')
const selectedImageModel = ref<string>('')

const hasScript = computed(() => {
  const currentEp = currentEpisode.value
  return currentEp && currentEp.script_content && currentEp.script_content.length > 0
})

const currentEpisode = computed(() => {
  if (!drama.value?.episodes) return null
  return drama.value.episodes.find(
    ep => ep.episode_number === episodeNumber
  )
})

const hasCharacters = computed(() => {
  return currentEpisode.value?.characters && currentEpisode.value.characters.length > 0
})

const charactersCount = computed(() => {
  return currentEpisode.value?.characters?.length || 0
})

const hasExtractedData = computed(() => {
  const hasScenes = currentEpisode.value?.scenes && currentEpisode.value.scenes.length > 0
  // 只要有角色或场景，就认为已经提取过数据
  return hasCharacters.value || hasScenes
})

const allImagesGenerated = computed(() => {
  // 如果没有提取任何数据，允许跳过（可能是空章节或用户想直接进入拆解分镜）
  if (!hasExtractedData.value) return true
  
  const characters = currentEpisode.value?.characters || []
  const scenes = currentEpisode.value?.scenes || []
  
  // 如果角色和场景都为空，允许跳过
  if (characters.length === 0 && scenes.length === 0) return true
  
  // 检查所有有数据的项是否都已生成图片
  const allCharsHaveImages = characters.length === 0 || characters.every(char => char.image_url)
  const allScenesHaveImages = scenes.length === 0 || scenes.every(scene => scene.image_url)
  
  return allCharsHaveImages && allScenesHaveImages
})

const goBack = () => {
  // 使用 replace 避免在历史记录中留下当前页面
  router.replace(`/dramas/${dramaId}`)
}

// 加载AI模型配置
const loadAIConfigs = async () => {
  try {
    const [textList, imageList] = await Promise.all([
      aiAPI.list('text'),
      aiAPI.list('image')
    ])
    
    // 只使用激活的配置
    const activeTextList = textList.filter(c => c.is_active)
    const activeImageList = imageList.filter(c => c.is_active)
    
    // 展开模型列表并去重（保留优先级最高的）
    const allTextModels = activeTextList.flatMap(config => {
      const models = Array.isArray(config.model) ? config.model : [config.model]
      return models.map(modelName => ({
        modelName,
        configName: config.name,
        configId: config.id,
        priority: config.priority || 0
      }))
    }).sort((a, b) => b.priority - a.priority)
    
    // 按模型名称去重，保留优先级最高的（已排序，第一个就是优先级最高的）
    const textModelMap = new Map<string, ModelOption>()
    allTextModels.forEach(model => {
      if (!textModelMap.has(model.modelName)) {
        textModelMap.set(model.modelName, model)
      }
    })
    textModels.value = Array.from(textModelMap.values())
    
    const allImageModels = activeImageList.flatMap(config => {
      const models = Array.isArray(config.model) ? config.model : [config.model]
      return models.map(modelName => ({
        modelName,
        configName: config.name,
        configId: config.id,
        priority: config.priority || 0
      }))
    }).sort((a, b) => b.priority - a.priority)
    
    // 按模型名称去重，保留优先级最高的
    const imageModelMap = new Map<string, ModelOption>()
    allImageModels.forEach(model => {
      if (!imageModelMap.has(model.modelName)) {
        imageModelMap.set(model.modelName, model)
      }
    })
    imageModels.value = Array.from(imageModelMap.values())
    
    // 设置默认选择（优先级最高的）
    if (textModels.value.length > 0 && !selectedTextModel.value) {
      selectedTextModel.value = textModels.value[0].modelName
    }
    if (imageModels.value.length > 0 && !selectedImageModel.value) {
      // 优先选择包含 nano 的模型
      const nanoModel = imageModels.value.find(m => m.modelName.toLowerCase().includes('nano'))
      selectedImageModel.value = nanoModel ? nanoModel.modelName : imageModels.value[0].modelName
    }
    
    // 验证已选择的模型是否还在可用列表中，如果不在则重置为默认值
    const availableTextModelNames = textModels.value.map(m => m.modelName)
    const availableImageModelNames = imageModels.value.map(m => m.modelName)
    
    if (selectedTextModel.value && !availableTextModelNames.includes(selectedTextModel.value)) {
      console.warn(`已选择的文本模型 ${selectedTextModel.value} 不在可用列表中，重置为默认值`)
      selectedTextModel.value = textModels.value.length > 0 ? textModels.value[0].modelName : ''
      // 更新 localStorage
      if (selectedTextModel.value) {
        localStorage.setItem(`ai_text_model_${dramaId}`, selectedTextModel.value)
      }
    }
    
    if (selectedImageModel.value && !availableImageModelNames.includes(selectedImageModel.value)) {
      console.warn(`已选择的图片模型 ${selectedImageModel.value} 不在可用列表中，重置为默认值`)
      // 优先选择包含 nano 的模型
      const nanoModel = imageModels.value.find(m => m.modelName.toLowerCase().includes('nano'))
      selectedImageModel.value = imageModels.value.length > 0 ? (nanoModel ? nanoModel.modelName : imageModels.value[0].modelName) : ''
      // 更新 localStorage
      if (selectedImageModel.value) {
        localStorage.setItem(`ai_image_model_${dramaId}`, selectedImageModel.value)
      }
    }
  } catch (error: any) {
    console.error('加载AI配置失败:', error)
  }
}

// 显示模型配置对话框
const showModelConfigDialog = () => {
  modelConfigDialogVisible.value = true
  loadAIConfigs()
}

// 保存模型配置
const saveModelConfig = () => {
  if (!selectedTextModel.value || !selectedImageModel.value) {
    ElMessage.warning($t('workflow.pleaseSelectModels'))
    return
  }
  
  // 保存模型名称到localStorage
  localStorage.setItem(`ai_text_model_${dramaId}`, selectedTextModel.value)
  localStorage.setItem(`ai_image_model_${dramaId}`, selectedImageModel.value)
  
  ElMessage.success($t('workflow.modelConfigSaved'))
  modelConfigDialogVisible.value = false
}

const nextStep = () => {
  if (currentStep.value < 3) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

// 从localStorage加载已保存的模型配置
const loadSavedModelConfig = () => {
  const savedTextModel = localStorage.getItem(`ai_text_model_${dramaId}`)
  const savedImageModel = localStorage.getItem(`ai_image_model_${dramaId}`)
  
  if (savedTextModel) {
    selectedTextModel.value = savedTextModel
  }
  if (savedImageModel) {
    selectedImageModel.value = savedImageModel
  }
}

const loadDramaData = async () => {
  try {
    const data = await dramaAPI.get(dramaId)
    drama.value = data
    
    if (!hasScript.value) {
      scriptContent.value = ''
      // 如果没有剧本内容，重置到第一步
      currentStep.value = 0
    }

    // 检查是否有生成中的角色或场景，自动启动轮询
    await checkAndStartPolling()
  } catch (error: any) {
    ElMessage.error(error.message || '加载项目数据失败')
  }
}

// 检查并启动轮询
const checkAndStartPolling = async () => {
  if (!currentEpisode.value) return

  // 检查角色的生成状态
  for (const char of currentEpisode.value.characters || []) {
    if (char.image_generation_status === 'pending' || char.image_generation_status === 'processing') {
      // 查找对应的image_generation记录
      try {
        const imageGenList = await imageAPI.listImages({
          drama_id: dramaId,
          status: char.image_generation_status as any
        })
        
        // 找到这个角色的image_generation记录
        const charImageGen = imageGenList.items.find(img => 
          img.character_id === char.id && (img.status === 'pending' || img.status === 'processing')
        )
        
        if (charImageGen) {
          // 启动轮询
          generatingCharacterImages.value[char.id] = true
          pollImageStatus(charImageGen.id, async () => {
            await loadDramaData()
            ElMessage.success(`${char.name}的图片生成完成！`)
          }).finally(() => {
            generatingCharacterImages.value[char.id] = false
          })
        }
      } catch (error) {
        console.error('[轮询] 查询角色图片生成记录失败:', error)
      }
    }
  }

  // 检查场景的生成状态
  for (const scene of currentEpisode.value.scenes || []) {
    if (scene.image_generation_status === 'pending' || scene.image_generation_status === 'processing') {
      // 查找对应的image_generation记录
      try {
        const imageGenList = await imageAPI.listImages({
          drama_id: dramaId,
          status: scene.image_generation_status as any
        })
        
        // 找到这个场景的image_generation记录
        const sceneImageGen = imageGenList.items.find(img => 
          img.scene_id === scene.id && (img.status === 'pending' || img.status === 'processing')
        )
        
        if (sceneImageGen) {
          // 启动轮询
          generatingSceneImages.value[scene.id] = true
          pollImageStatus(sceneImageGen.id, async () => {
            await loadDramaData()
            ElMessage.success(`${scene.location}的图片生成完成！`)
          }).finally(() => {
            generatingSceneImages.value[scene.id] = false
          })
        }
      } catch (error) {
        console.error('[轮询] 查询场景图片生成记录失败:', error)
      }
    }
  }
}

const saveChapterScript = async () => {
  try {
    const existingEpisodes = drama.value?.episodes || []
    
    // 查找当前章节
    const episodeIndex = existingEpisodes.findIndex(
      ep => ep.episode_number === episodeNumber
    )
    
    let updatedEpisodes
    if (episodeIndex >= 0) {
      // 更新已有章节
      updatedEpisodes = [...existingEpisodes]
      updatedEpisodes[episodeIndex] = {
        ...updatedEpisodes[episodeIndex],
        script_content: scriptContent.value
      }
    } else {
      // 创建新章节
      const newEpisode = {
        episode_number: episodeNumber,
        title: `第${episodeNumber}集`,
        script_content: scriptContent.value
      }
      updatedEpisodes = [...existingEpisodes, newEpisode]
    }
    
    await dramaAPI.saveEpisodes(dramaId, updatedEpisodes)
    ElMessage.success('章节保存成功！')
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  }
}

const editCurrentEpisodeScript = () => {
  scriptContent.value = currentEpisode.value?.script_content || ''
}

const handleExtractCharactersAndBackgrounds = async () => {
  // 如果已经提取过，显示确认对话框
  if (hasExtractedData.value) {
    try {
      await ElMessageBox.confirm(
        $t('workflow.reExtractConfirmMessage'),
        $t('workflow.reExtractConfirmTitle'),
        {
          confirmButtonText: $t('common.confirm'),
          cancelButtonText: $t('common.cancel'),
          type: 'warning',
          distinguishCancelAndClose: true
        }
      )
    } catch {
      ElMessage.info('已取消提取')
      return
    }
  }
  
  // 显示即将开始的提示
  if (hasExtractedData.value) {
    ElMessage.info($t('workflow.startReExtracting'))
  }
  
  await extractCharactersAndBackgrounds()
}

// 轮询检查图片生成状态
const pollImageStatus = async (imageGenId: number, onComplete: () => Promise<void>) => {
  const maxAttempts = 100 // 最多轮询100次
  const pollInterval = 6000 // 每6秒轮询一次
  
  for (let i = 0; i < maxAttempts; i++) {
    try {
      await new Promise(resolve => setTimeout(resolve, pollInterval))
      
      const imageGen = await imageAPI.getImage(imageGenId)
      
      if (imageGen.status === 'completed') {
        // 生成成功
        await onComplete()
        return
      } else if (imageGen.status === 'failed') {
        // 生成失败
        ElMessage.error(`图片生成失败: ${imageGen.error_msg || '未知错误'}`)
        return
      }
      // 如果是pending或processing，继续轮询
    } catch (error: any) {
      console.error('[轮询] 检查图片状态失败:', error)
      // 继续轮询，不中断
    }
  }
  
  // 超时
  ElMessage.warning('图片生成超时，请稍后刷新页面查看结果')
}

const extractCharactersAndBackgrounds = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error('章节信息不存在')
    return
  }

  extractingCharactersAndBackgrounds.value = true
  
  try {
    const episodeId = currentEpisode.value.id

    // 并行创建异步任务
    const [characterTask, backgroundTask] = await Promise.all([
      generationAPI.generateCharacters({
        drama_id: dramaId.toString(),
        episode_id: parseInt(episodeId),
        outline: currentEpisode.value.script_content || '',
        count: 0,
        model: selectedTextModel.value  // 传递用户选择的文本模型
      }),
      dramaAPI.extractBackgrounds(episodeId.toString(), selectedTextModel.value)  // 传递用户选择的文本模型
    ])
    
    ElMessage.success('任务已创建，正在后台处理...')
    
    // 并行轮询两个任务
    await Promise.all([
      pollExtractTask(characterTask.task_id, 'character'),
      pollExtractTask(backgroundTask.task_id, 'background')
    ])
    
    ElMessage.success('角色和场景提取成功！')
    await loadDramaData()
  } catch (error: any) {
    console.error('角色和场景提取失败:', error)
    
    const errorData = error.response?.data?.error
    const errorMsg = errorData?.message || error.message || '提取失败'
    
    if (errorMsg.includes('no config found') || 
        errorMsg.includes('AI client') ||
        errorMsg.includes('failed to get AI client')) {
      ElMessage({
        type: 'warning',
        message: '未配置AI服务，请前往"设置 > AI服务配置"添加文本生成服务',
        duration: 5000,
        showClose: true
      })
    } else {
      ElMessage.error(errorMsg)
    }
  } finally {
    extractingCharactersAndBackgrounds.value = false
  }
}

// 轮询提取任务状态
const pollExtractTask = async (taskId: string, type: 'character' | 'background') => {
  const maxAttempts = 300 // 最多轮询300次（10分钟）
  const interval = 2000 // 每2秒查询一次
  
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(resolve => setTimeout(resolve, interval))
    
    try {
      const task = await generationAPI.getTaskStatus(taskId)
      
      if (task.status === 'completed') {
        // 任务完成
        if (type === 'character' && task.result) {
          // 解析角色数据并保存
          const result = typeof task.result === 'string' ? JSON.parse(task.result) : task.result
          if (result.characters && result.characters.length > 0) {
            await dramaAPI.saveCharacters(dramaId, result.characters, currentEpisode.value?.id)
          }
        }
        return
      } else if (task.status === 'failed') {
        // 任务失败
        throw new Error(task.error || `${type === 'character' ? '角色生成' : '场景提取'}失败`)
      }
      // 否则继续轮询
    } catch (error: any) {
      console.error(`轮询${type}任务状态失败:`, error)
      throw error
    }
  }
  
  throw new Error(`${type === 'character' ? '角色生成' : '场景提取'}超时`)
}


const generateCharacterImage = async (characterId: number, styleParam?: string, sizeParam?: string, referenceWorkParam?: string, whiteBackgroundParam?: boolean) => {
  generatingCharacterImages.value[characterId] = true
  
  try {
    // 准备参数
    let style = styleParam
    let reference = referenceWorkParam
    let size = sizeParam
    
    // 如果未提供参数，则使用默认值
    if (!style && drama.value?.style) {
        style = drama.value.style
    }
    
    if (!reference && drama.value?.reference_work) {
        reference = drama.value.reference_work
    }
    
    if (!size) {
        if (drama.value?.aspect_ratio === '9:16') {
            size = '1440x2560'
        } else {
            size = '2560x1440'
        }
    }
    
    // 获取用户选择的图片生成模型
    const model = selectedImageModel.value || undefined
    const response = await characterLibraryAPI.generateCharacterImage(characterId.toString(), model, style, size, reference, whiteBackgroundParam)
    const imageGenId = response.image_generation?.id
    
    if (imageGenId) {
      ElMessage.info('角色图片生成中，请稍候...')
      // 轮询检查生成状态
      await pollImageStatus(imageGenId, async () => {
        await loadDramaData()
        ElMessage.success('角色图片生成完成！')
      })
    } else {
      ElMessage.success('角色图片生成已启动')
      await loadDramaData()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
  } finally {
    generatingCharacterImages.value[characterId] = false
  }
}

const toggleSelectAllCharacters = () => {
  if (selectAllCharacters.value) {
    selectedCharacterIds.value = currentEpisode.value?.characters?.map(char => char.id) || []
  } else {
    selectedCharacterIds.value = []
  }
}

const toggleSelectAllScenes = () => {
  if (selectAllScenes.value) {
    selectedSceneIds.value = currentEpisode.value?.scenes?.map(scene => scene.id) || []
  } else {
    selectedSceneIds.value = []
  }
}

const batchGenerateCharacterImages = async () => {
  if (selectedCharacterIds.value.length === 0) {
    ElMessage.warning('请先选择要生成的角色')
    return
  }
  
  batchGeneratingCharacters.value = true
  try {
    // 获取用户选择的图片生成模型
    const model = selectedImageModel.value || undefined
    
    // 使用批量生成API
    await characterLibraryAPI.batchGenerateCharacterImages(
      selectedCharacterIds.value.map(id => id.toString()),
      model
    )
    
    ElMessage.success($t('workflow.batchTaskSubmitted'))
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || $t('workflow.batchGenerateFailed'))
  } finally {
    batchGeneratingCharacters.value = false
  }
}

const generateSceneImage = async (sceneId: number) => {
  generatingSceneImages.value[sceneId] = true
  
  try {
    // 获取用户选择的图片生成模型
    const model = selectedImageModel.value || undefined
    const response = await dramaAPI.generateSceneImage({ 
      scene_id: sceneId,
      model
    })
    const imageGenId = response.image_generation?.id
    
    if (imageGenId) {
      ElMessage.info($t('workflow.sceneImageGenerating'))
      // 轮询检查生成状态
      await pollImageStatus(imageGenId, async () => {
        await loadDramaData()
        ElMessage.success($t('workflow.sceneImageComplete'))
      })
    } else {
      ElMessage.success($t('workflow.sceneImageStarted'))
      await loadDramaData()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
  } finally {
    generatingSceneImages.value[sceneId] = false
  }
}

const batchGenerateSceneImages = async () => {
  if (selectedSceneIds.value.length === 0) {
    ElMessage.warning('请先选择要生成的场景')
    return
  }
  
  batchGeneratingScenes.value = true
  try {
    const promises = selectedSceneIds.value.map(sceneId => 
      generateSceneImage(sceneId)
    )
    const results = await Promise.allSettled(promises)
    
    const successCount = results.filter(r => r.status === 'fulfilled').length
    const failCount = results.filter(r => r.status === 'rejected').length
    
    if (failCount === 0) {
      ElMessage.success($t('workflow.batchCompleteSuccess', { count: successCount }))
    } else {
      ElMessage.warning($t('workflow.batchCompletePartial', { success: successCount, fail: failCount }))
    }
  } catch (error: any) {
    ElMessage.error(error.message || $t('workflow.batchGenerateFailed'))
  } finally {
    batchGeneratingScenes.value = false
  }
}

const taskProgress = ref(0)
const taskMessage = ref('')
let pollTimer: any = null

const generateShots = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error('章节信息不存在')
    return
  }
  
  generatingShots.value = true
  taskProgress.value = 0
  taskMessage.value = '初始化任务...'
  
  try {
    const episodeId = currentEpisode.value.id.toString()
    
    // 【调试日志】输出当前操作的集数信息
    console.log('=== 开始生成分镜 ===')
    console.log('当前 episodeNumber (路由参数):', episodeNumber)
    console.log('当前 episodeId (从 currentEpisode 获取):', episodeId)
    console.log('currentEpisode 完整信息:', {
      id: currentEpisode.value?.id,
      episode_number: currentEpisode.value?.episode_number,
      title: currentEpisode.value?.title
    })
    console.log('所有剧集列表:', drama.value?.episodes?.map(ep => ({ id: ep.id, episode_number: ep.episode_number, title: ep.title })))
    
    // 创建异步任务
    const response = await generationAPI.generateStoryboard(episodeId, selectedTextModel.value)
    
    taskMessage.value = response.message || '任务已创建'
    
    // 开始轮询任务状态
    await pollTaskStatus(response.task_id)
    
  } catch (error: any) {
    ElMessage.error(error.message || '拆分失败')
    generatingShots.value = false
  }
}

const pollTaskStatus = async (taskId: string) => {
  const checkStatus = async () => {
    try {
      const task = await generationAPI.getTaskStatus(taskId)
      
      taskProgress.value = task.progress
      taskMessage.value = task.message || `处理中... ${task.progress}%`
      
      if (task.status === 'completed') {
        // 任务完成
        if (pollTimer) {
          clearInterval(pollTimer)
          pollTimer = null
        }
        generatingShots.value = false
        
        ElMessage.success($t('workflow.splitSuccess'))
        
        // 跳转到专业编辑器页面
        router.push({
          name: 'ProfessionalEditor',
          params: {
            dramaId: dramaId,
            episodeNumber: episodeNumber
          }
        })
      } else if (task.status === 'failed') {
        // 任务失败
        if (pollTimer) {
          clearInterval(pollTimer)
          pollTimer = null
        }
        generatingShots.value = false
        ElMessage.error(task.error || '分镜拆分失败')
      }
      // 否则继续轮询
    } catch (error: any) {
      if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
      }
      generatingShots.value = false
      ElMessage.error('查询任务状态失败: ' + error.message)
    }
  }
  
  // 立即检查一次
  await checkStatus()
  
  // 每2秒轮询一次
  pollTimer = setInterval(checkStatus, 2000)
}

const regenerateShots = async () => {
  await ElMessageBox.confirm($t('workflow.reSplitConfirm'), $t('common.tip'), {
    type: 'warning'
  })
  
  await generateShots()
}

const shotEditDialogVisible = ref(false)
const editingShot = ref<any>(null)
const editingShotIndex = ref<number>(-1)
const savingShot = ref(false)

const editShot = (shot: any, index: number) => {
  editingShot.value = { ...shot }
  editingShotIndex.value = index
  shotEditDialogVisible.value = true
}

const saveShotEdit = async () => {
  if (!editingShot.value) return
  
  try {
    savingShot.value = true
    
    // 调用API更新镜头
    await dramaAPI.updateStoryboard(editingShot.value.id, editingShot.value)
    
    // 更新本地数据
    if (currentEpisode.value?.storyboards) {
      currentEpisode.value.storyboards[editingShotIndex.value] = { ...editingShot.value }
    }
    
    ElMessage.success('镜头修改成功')
    shotEditDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
  } finally {
    savingShot.value = false
  }
}

// 对话框相关方法
const openPromptDialog = (item: any, type: 'character' | 'scene') => {
  currentEditItem.value = item
  currentEditItem.value.name = item.name || item.location
  currentEditType.value = type
  editPrompt.value = item.prompt || item.appearance || item.description || ''
  // 重置优化历史，保证每次打开弹窗可独立撤销
  optimizePromptHistory.value = ''
  canUndoOptimize.value = false
  
  // 初始化额外字段
  editStyle.value = drama.value?.style || ''
  editReference.value = drama.value?.reference_work || ''
  
  // 确保风格和参考作品在提示词中可见
  if (editStyle.value && !editPrompt.value.toLowerCase().includes(editStyle.value.toLowerCase())) {
    editPrompt.value += `, ${editStyle.value} style`
  }
  
  if (editReference.value && !editPrompt.value.toLowerCase().includes(editReference.value.toLowerCase())) {
    editPrompt.value += `, style reference: ${editReference.value}`
  }
  
  // 根据宽高比设置默认尺寸
  if (drama.value?.aspect_ratio === '9:16') {
      editSize.value = '1440x2560'
  } else {
      editSize.value = '2560x1440'
  }
  
  // 初始化视图设置
  if (type === 'character') {
      const settings = []
      // 默认选中，如果提示词中没有明确否定（这里简单处理为默认选中）
      // 检查提示词中是否已包含
      const currentPrompt = editPrompt.value.toLowerCase()
      
      // 白底图
      // 检查提示词中是否包含白底相关的关键词
      const whiteBgKeywords = ['white background', 'simple background', 'solid background', 'blank background']
      const hasWhiteBg = whiteBgKeywords.some(keyword => currentPrompt.includes(keyword))
      
      if (hasWhiteBg || !currentPrompt) {
          settings.push('white_background')
      }
      
      // 三视图
      // 检查提示词中是否包含三视图相关的关键词
      const threeViewsKeywords = ['three views', 'front view', 'side view', 'back view', 'character turnaround', 'character sheet', 'orthographic']
      const hasThreeViews = threeViewsKeywords.some(keyword => currentPrompt.includes(keyword))
      
      if (hasThreeViews || !currentPrompt) {
          settings.push('three_views')
      }
      
      // 构图修正 (默认选中)
      // 检查是否包含任何一个构图修正的关键正向词
      const hasCompositionPrompts = ['best quality', 'character centered', 'full body fully visible', 'whole body', 'no cropping', 'padding around edges']
          .some(p => currentPrompt.includes(p))
      
      if (hasCompositionPrompts || !currentPrompt) {
          settings.push('composition_fix')
      }
      
      editViewSettings.value = settings
  } else {
      editViewSettings.value = []
  }
  
  promptDialogVisible.value = true
}

// 保护性提示词识别与优化处理
const optimizePrompt = () => {
  // 提示词为空时不进行优化
  if (!editPrompt.value || !editPrompt.value.trim()) {
    ElMessage.warning('提示词为空，无法优化')
    return
  }
  // 记录优化前内容用于撤销
  optimizePromptHistory.value = editPrompt.value
  const optimized = buildOptimizedPrompt(editPrompt.value)
  editPrompt.value = optimized
  canUndoOptimize.value = true
  ElMessage.success('提示词已优化，已保护白底、三视图、基础风格及构图关键词')
}

// 撤销优化提示词
const undoOptimizePrompt = () => {
  if (!canUndoOptimize.value || !optimizePromptHistory.value) {
    return
  }
  editPrompt.value = optimizePromptHistory.value
  optimizePromptHistory.value = ''
  canUndoOptimize.value = false
  ElMessage.info('已撤销优化')
}

// 构建优化后的提示词（不修改保护段落）
const buildOptimizedPrompt = (rawPrompt: string) => {
  // 规范化标点与空白
  const normalized = rawPrompt
    .replace(/，/g, ',')
    .replace(/\s+/g, ' ')
    .replace(/,+/g, ',')
    .trim()

  const segments = normalized
    .split(/[,，\n]+/)
    .map(item => item.trim())
    .filter(Boolean)

  // 白底图保护关键词
  const whiteBgProtected = [
    '100% pure white solid background',
    'entire image is solid white',
    'no black bars',
    'no black borders',
    'no dark areas',
    'completely blank background',
    'flat white background',
    'no scenery',
    'no environmental elements',
    'no extra details outside the character',
    'white background',
    'solid white background',
    'pure white background',
    '纯白背景',
    '白底'
  ]
  // 三视图保护关键词
  const threeViewsProtected = [
    'character turnaround sheet',
    'orthographic design drawing',
    'three orthogonal views',
    'front view',
    'side view',
    'back view',
    'no missing views',
    'consistent character details',
    'professional character design sheet',
    '三视图',
    '正面',
    '侧面',
    '背面'
  ]
  // 构图修正保护关键词
  const compositionProtected = [
    'full body fully visible',
    'whole body',
    'no cropping',
    'no out of frame',
    'no partial body',
    'padding around edges',
    'character centered',
    'clean and tidy layout',
    'no ugly',
    'no disfigured',
    'no malformed limbs',
    'no extra limbs',
    'no missing limbs',
    '中心构图',
    '对称',
    '黄金分割'
  ]
  // 基础画面风格保护关键词
  const styleCoreKeywords = [
    'cel-shaded anime',
    'makoto shinkai',
    'studio ghibli',
    'american animation',
    'chinese animation',
    'shanghai ink',
    'impasto fantasy',
    'magical girl',
    'sci-fi cyberpunk',
    'chibi',
    'light novel cover',
    'realistic',
    '写实',
    '赛璐璐',
    '新海诚',
    '吉卜力',
    '欧美',
    '国漫',
    '水墨',
    '厚涂',
    '魔法少女',
    '赛博',
    'Q版'
  ]
  // 参考作品保护关键词
  const referenceProtected = ['style reference', '参考作品', 'reference work']
  // 白底图冲突关键词（仅用于非保护段落清理）
  const whiteBgConflict = [
    'dark background',
    'black background',
    'gray background',
    'grey background',
    'gradient background',
    'vignette',
    'shadow',
    'floor',
    'ground',
    'scenery',
    'environment'
  ]

  // 判断是否有白底保护段落
  const hasWhiteBgProtected = segments.some(seg => {
    const lower = seg.toLowerCase()
    return whiteBgProtected.some(keyword => lower.includes(keyword))
  })

  const resultSegments: string[] = []
  const seen = new Set<string>()
  let hasStyleCore = false

  // 逐段处理，保护段落完全保留
  segments.forEach(seg => {
    const lower = seg.toLowerCase()
    const isWhiteProtected = whiteBgProtected.some(keyword => lower.includes(keyword))
    const isThreeViewsProtected = threeViewsProtected.some(keyword => lower.includes(keyword))
    const isCompositionProtected = compositionProtected.some(keyword => lower.includes(keyword))
    const isStyleProtected = styleCoreKeywords.some(keyword => lower.includes(keyword))
    const isReferenceProtected = referenceProtected.some(keyword => lower.includes(keyword))

    if (isStyleProtected) {
      hasStyleCore = true
    }

    const normalizedKey = lower
    if (seen.has(normalizedKey)) {
      return
    }

    // 保护段落直接保留
    if (isWhiteProtected || isThreeViewsProtected || isCompositionProtected || isStyleProtected || isReferenceProtected) {
      resultSegments.push(seg)
      seen.add(normalizedKey)
      return
    }

    // 清理与白底冲突的非保护段落
    if (hasWhiteBgProtected && whiteBgConflict.some(keyword => lower.includes(keyword))) {
      return
    }

    // 清理多余空白
    const cleaned = seg.replace(/\s+/g, ' ').trim()
    if (!cleaned) return

    resultSegments.push(cleaned)
    seen.add(normalizedKey)
  })

  // 适度增强基础风格描述（仅补充，避免显著增量）
  const hasStyleEnhancement = resultSegments.some(seg =>
    /cinematic|highly detailed|细节|电影质感/i.test(seg)
  )
  if (hasStyleCore && !hasStyleEnhancement && rawPrompt.length < 400) {
    resultSegments.push('highly detailed, clean silhouette')
  }

  return resultSegments.join(', ')
}

const savePrompt = async () => {
  try {
    if (currentEditType.value === 'character') {
      // 处理视图设置，追加到提示词
      let finalPrompt = editPrompt.value
      
      if (editViewSettings.value.includes('white_background')) {
          // 1. 白底图提示词
          const whiteBgPrompts = [
              '100% pure white solid background',
              'entire image is solid white',
              'no black bars',
              'no black borders',
              'no dark areas',
              'completely blank background',
              'flat white background',
              'no scenery',
              'no environmental elements',
              'no extra details outside the character'
          ]
          
          const missingBgPrompts = whiteBgPrompts.filter(p => !finalPrompt.toLowerCase().includes(p))
          if (missingBgPrompts.length > 0) {
              finalPrompt += ', ' + missingBgPrompts.join(', ')
          }
      }
      
      if (editViewSettings.value.includes('three_views')) {
           // 2. 三视图提示词
           const views = [
               'character turnaround sheet',
               'orthographic design drawing',
               'three orthogonal views of the same character on one single image',
               'front view', 
               'side view', 
               'back view',
               'no missing views',
               'consistent character details across all views',
               'professional character design sheet'
           ]
           const missingViews = views.filter(v => !finalPrompt.toLowerCase().includes(v))
           if (missingViews.length > 0) {
               finalPrompt += ', ' + missingViews.join(', ')
           }
       }
       
       if (editViewSettings.value.includes('composition_fix')) {
           // 3. 构图修正提示词
           const compositionPrompts = [
               'full body fully visible',
               'whole body',
               'no cropping',
               'no out of frame',
               'no partial body',
               'padding around edges',
               'character centered',
               'clean and tidy layout',
               'no ugly',
               'no disfigured',
               'no malformed limbs',
               'no extra limbs',
               'no missing limbs'
           ]
           
           const missingPrompts = compositionPrompts.filter(p => !finalPrompt.toLowerCase().includes(p))
           if (missingPrompts.length > 0) {
               finalPrompt += ', ' + missingPrompts.join(', ')
           }
       }
       
       // 更新提示词显示
       editPrompt.value = finalPrompt
      
      await characterLibraryAPI.updateCharacter(currentEditItem.value.id, {
        appearance: finalPrompt
      })
      
      // 计算是否启用白底图
      const useWhiteBackground = editViewSettings.value.includes('white_background')
      
      // 传递自定义参数进行生成
      await generateCharacterImage(currentEditItem.value.id, editStyle.value, editSize.value, editReference.value, useWhiteBackground)
    } else {
      // 1. 先保存场景提示词
      await dramaAPI.updateScenePrompt(currentEditItem.value.id.toString(), editPrompt.value)
      
      // 2. 生成场景图片
      const model = selectedImageModel.value || undefined
      const response = await dramaAPI.generateSceneImage({ 
        scene_id: parseInt(currentEditItem.value.id),
        prompt: editPrompt.value,
        model
      })
      const imageGenId = response.image_generation?.id
      
      // 3. 轮询图片生成状态
      if (imageGenId) {
        ElMessage.info('场景图片生成中，请稍候...')
        generatingSceneImages.value[currentEditItem.value.id] = true
        pollImageStatus(imageGenId, async () => {
          await loadDramaData()
          ElMessage.success('场景图片生成完成！')
        }).finally(() => {
          generatingSceneImages.value[currentEditItem.value.id] = false
        })
      } else {
        ElMessage.success('场景图片生成已启动')
        await loadDramaData()
      }
    }
    promptDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  }
}

const uploadCharacterImage = (characterId: number) => {
  currentUploadTarget.value = { id: characterId, type: 'character' }
  uploadDialogVisible.value = true
}

const uploadSceneImage = (sceneId: number) => {
  currentUploadTarget.value = { id: sceneId, type: 'scene' }
  uploadDialogVisible.value = true
}

const selectFromLibrary = async (characterId: number) => {
  try {
    const result = await characterLibraryAPI.list({ page_size: 50 })
    libraryItems.value = result.items || []
    currentUploadTarget.value = characterId
    libraryDialogVisible.value = true
  } catch (error: any) {
    ElMessage.error(error.message || $t('workflow.loadLibraryFailed'))
  }
}

const addToCharacterLibrary = async (character: any) => {
  if (!character.image_url) {
    ElMessage.warning($t('workflow.generateImageFirst'))
    return
  }
  
  try {
    await ElMessageBox.confirm(
      $t('workflow.addToLibraryConfirm', { name: character.name }),
      $t('workflow.addToLibrary'),
      {
        confirmButtonText: $t('common.confirm'),
        cancelButtonText: $t('common.cancel'),
        type: 'info'
      }
    )
    
    await characterLibraryAPI.addCharacterToLibrary(character.id.toString())
    ElMessage.success($t('workflow.addedToLibrary'))
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || $t('workflow.addFailed'))
    }
  }
}

const selectLibraryItem = async (item: any) => {
  try {
    if (currentUploadTarget.value?.type === 'character') {
      await characterLibraryAPI.applyFromLibrary(
        currentUploadTarget.value.id.toString(),
        item.id
      )
      ElMessage.success('应用角色形象成功！')
      await loadDramaData()
      libraryDialogVisible.value = false
    }
  } catch (error: any) {
    ElMessage.error(error.message || '应用失败')
  }
}

const handleUploadSuccess = async (response: any) => {
  try {
    const imageUrl = response.url || response.data?.url
    if (!imageUrl) {
      ElMessage.error('上传失败：未获取到图片地址')
      return
    }

    if (currentUploadTarget.value?.type === 'character') {
      await characterLibraryAPI.uploadCharacterImage(
        currentUploadTarget.value.id.toString(),
        imageUrl
      )
      ElMessage.success('上传成功！')
    } else if (currentUploadTarget.value?.type === 'scene') {
      // TODO: 场景图片上传API
      ElMessage.success('上传成功！')
    }
    
    await loadDramaData()
    uploadDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  }
}

const handleUploadError = () => {
  ElMessage.error('上传失败，请重试')
}

const deleteCharacter = async (characterId: number) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该角色吗？删除后将无法恢复。',
      '删除确认',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )

    await characterLibraryAPI.deleteCharacter(characterId)
    ElMessage.success('角色已删除')
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const goToProfessionalUI = () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error('章节信息不存在')
    return
  }
  
  router.push({
    name: 'ProfessionalEditor',
    params: {
      dramaId: dramaId,
      episodeNumber: episodeNumber
    }
  })
}

const goToCompose = () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error('章节信息不存在')
    return
  }
  
  router.push({
    name: 'SceneComposition',
    params: {
      id: dramaId,
      episodeId: currentEpisode.value.id
    }
  })
}

// 监听步骤变化，保存到 localStorage
watch(currentStep, (newStep) => {
  localStorage.setItem(getStepStorageKey(), newStep.toString())
})

onMounted(() => {
  loadDramaData()
  loadSavedModelConfig()
  loadAIConfigs()
})
</script>

<style scoped lang="scss">
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  // padding: var(--space-2) var(--space-3);
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    // padding: var(--space-3) var(--space-4);
  }
}

@media (min-width: 1024px) {
  .page-container {
    // padding: var(--space-4) var(--space-5);
  }
}

.content-wrapper {
  margin: 0 auto;
  width: 100%;
}

.project-info-section {
  margin-bottom: var(--space-3);
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border-primary);
  
  :deep(.el-descriptions__label) {
    background-color: var(--bg-secondary);
    font-weight: 500;
    color: var(--text-secondary);
  }
  
  :deep(.el-descriptions__content) {
    background-color: var(--bg-card);
    color: var(--text-primary);
  }
}

/* Header styles matching PageHeader component */
.page-header {
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-primary);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;

  &:hover {
    background: var(--bg-card-hover);
    color: var(--text-primary);
    border-color: var(--border-secondary);
  }
}

.nav-divider {
  width: 1px;
  height: 2rem;
  background: var(--border-primary);
}

.header-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.025em;
  line-height: 1.2;
  white-space: nowrap;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  flex-shrink: 0;
}

.workflow-card {
  margin: 12px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-primary);

  :deep(.el-card__body) {
    padding: 0;
  }
}

.custom-steps {
  display: flex;
  align-items: center;
  gap: 12px;

  .step-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 20px;
    background: var(--bg-card-hover);
    transition: all 0.3s;

    &.active {
      background: var(--accent-light);
      
      .step-circle {
        background: var(--accent);
        color: var(--text-inverse);
      }
    }

    &.current {
      background: var(--accent);
      color: var(--text-inverse);
      
      .step-circle {
        background: var(--bg-card);
        color: var(--accent);
      }
      
      .step-text {
        color: var(--text-inverse);
      }
    }

    .step-circle {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--border-secondary);
      color: var(--text-secondary);
      font-weight: 600;
      transition: all 0.3s;
    }

    .step-text {
      font-size: 14px;
      font-weight: 500;
      white-space: nowrap;
    }
  }

  .step-arrow {
    color: var(--border-secondary);
  }
}

.stage-card {
  margin: 12px;
  
  &.stage-card-fullscreen {
    .stage-body-fullscreen {
      min-height: calc(100vh - 200px);
    }
  }
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .header-info {
      h2 {
        margin: 0 0 4px 0;
        font-size: 20px;
      }

      p {
        margin: 0;
        color: var(--text-muted);
        font-size: 14px;
      }
    }
  }
}

.stage-body {
  background: var(--bg-card);
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin: 12px 0;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
}

.action-buttons-inline {
  display: flex;
  gap: 12px;
}

.script-textarea {
  margin: 16px 0;
  
  &.script-textarea-fullscreen {
    :deep(textarea) {
      min-height: 500px;
      font-size: 14px;
      line-height: 1.8;
    }
  }
}

.image-gen-section {
  margin-bottom: 32px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-secondary);
    // border-radius: 8px;
    // border: 1px solid var(--border-primary);

    .section-title {
      display: flex;
      align-items: center;
      gap: 16px;

      h3 {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0;
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);

        .el-icon {
          color: var(--accent);
          font-size: 18px;
        }
      }

      .el-alert {
        border-radius: 4px;
      }
    }

    .section-actions {
      display: flex;
      align-items: center;
    }
  }
}


.empty-shots {
  padding: 60px 0;
  text-align: center;
}

.extracted-title {
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.secondary-text {
  color: var(--text-muted);
  margin-left: 4px;
}

.task-message {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
}

.model-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.fixed-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-primary);
  box-shadow: var(--shadow-card);
  transition: all 0.2s;

  &:hover {
    box-shadow: var(--shadow-card-hover);
  }

  :deep(.el-card__body) {
    flex: 1;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  .card-header {
    padding: 14px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-primary);
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-left {
      flex: 1;
      min-width: 0;

      h4 {
        margin: 0 0 4px 0;
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .el-tag {
        margin-top: 0;
      }
    }
  }

  .card-image-container {
    flex: 1;
    width: 100%;
    min-height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-secondary);

    .char-image,
    .scene-image {
      width: 100%;
      height: 100%;
      position: relative;
      z-index: 1;

      .el-image {
        width: 100%;
        height: 100%;
        border-radius: 0;
      }
    }

    .char-placeholder,
    .scene-placeholder {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: var(--text-muted);
      padding: 20px;
      
      &.generating {
        color: var(--warning);
        background: var(--warning-light);
        
        .rotating {
          animation: rotating 2s linear infinite;
        }
      }
      
      &.failed {
        color: var(--error);
        background: var(--error-light);
      }
      position: relative;
      z-index: 1;

      .el-icon {
        opacity: 0.5;
      }

      span {
        margin-top: 10px;
        font-size: 12px;
      }
    }
  }

  .card-actions {
    padding: 10px;
    background: var(--bg-card);
    border-top: 1px solid var(--border-primary);
    display: flex;
    justify-content: center;
    gap: 8px;

    .el-button {
      margin: 0;
    }
  }
}

.character-image-list,
.scene-image-list {
  padding: 5px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  margin-top: 16px;

  .character-item,
  .scene-item {
    min-height: 360px;
  }
}

// 角色库选择对话框
.library-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;
  padding: 8px;

  .library-item {
    cursor: pointer;
    border: 2px solid transparent;
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.3s;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-lg);
    }

    .el-image {
      width: 100%;
      height: 150px;
    }

    .library-item-name {
      padding: 8px;
      text-align: center;
      font-size: 12px;
      background: var(--bg-secondary);
      color: var(--text-primary);
    }
  }
}

.empty-library {
  padding: 40px 0;
}

// 上传区域
.upload-area {
  :deep(.el-upload-dragger) {
    width: 100%;
    height: 200px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
}

// 旋转动画
@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* ========================================
   Dark Mode / 深色模式
   ======================================== */
:deep(.el-card) {
  background: var(--bg-card);
  border-color: var(--border-primary);
}

:deep(.el-card__header) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}

:deep(.el-table) {
  --el-table-bg-color: var(--bg-card);
  --el-table-header-bg-color: var(--bg-secondary);
  --el-table-tr-bg-color: var(--bg-card);
  --el-table-row-hover-bg-color: var(--bg-card-hover);
  --el-table-border-color: var(--border-primary);
  --el-table-text-color: var(--text-primary);
  background: var(--bg-card);
}

:deep(.el-table th.el-table__cell),
:deep(.el-table td.el-table__cell) {
  background: var(--bg-card);
  border-color: var(--border-primary);
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell) {
  background: var(--bg-secondary);
}

:deep(.el-table__header-wrapper th) {
  background: var(--bg-secondary) !important;
  color: var(--text-secondary);
}

:deep(.el-dialog) {
  background: var(--bg-card);
}

:deep(.el-dialog__header) {
  background: var(--bg-card);
}

:deep(.el-form-item__label) {
  color: var(--text-primary);
}

:deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-input__inner) {
  color: var(--text-primary);
}

:deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  color: var(--text-primary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-select-dropdown) {
  background: var(--bg-elevated);
  border-color: var(--border-primary);
}

:deep(.el-upload-dragger) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}

.project-info-section {
  background: var(--bg-card);
  padding: 16px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  margin-bottom: 20px;

  .info-descriptions {
    :deep(.el-descriptions__label) {
      font-weight: bold;
      color: var(--text-secondary);
      background-color: var(--bg-secondary);
    }
    :deep(.el-descriptions__content) {
      color: var(--text-primary);
      background-color: var(--bg-card);
    }
  }
}
</style>
