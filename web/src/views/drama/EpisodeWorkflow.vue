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
                  <el-tag v-if="hasProps" type="success">{{ $t('workflow.props') }}: {{ propsCount }}</el-tag>
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

            <div v-if="hasProps">
              <h4 class="extracted-title">{{ $t('workflow.extractedProps') }}：</h4>
              <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                <el-tag 
                  v-for="prop in propsForEpisode" 
                  :key="prop.id"
                  type="primary"
                >
                  {{ prop.name }} <span v-if="prop.type" class="secondary-text">({{ prop.type }})</span>
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
                  <el-tooltip content="引用已有素材" placement="top">
                    <el-button 
                      size="small" 
                      @click="openAssetReference('character', char)"
                      :icon="More"
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
                  <el-tooltip content="引用已有素材" placement="top">
                    <el-button 
                      size="small" 
                      @click="openAssetReference('scene', scene)"
                      :icon="More"
                      circle
                    />
                  </el-tooltip>
                </div>
              </el-card>
            </div>
          </div>
        </div>

        <el-divider />

        <!-- 道具图片生成 -->
        <div class="image-gen-section">
          <div class="section-header">
            <div class="section-title">
              <h3>
                <el-icon><Box /></el-icon>
                {{ $t('workflow.propImages') }}
              </h3>
              <el-alert 
                type="info"
                :closable="false"
                style="margin: 0;"
              >
                {{ $t('workflow.propCount', { count: propsCount }) }}
              </el-alert>
            </div>
            <div class="section-actions">
              <el-checkbox 
                v-model="selectAllProps"
                @change="toggleSelectAllProps"
                style="margin-right: 12px;"
              >
                {{ $t('workflow.selectAll') }}
              </el-checkbox>
              <el-button 
                type="primary"
                @click="batchGeneratePropImages"
                :disabled="selectedPropIds.length === 0"
                size="default"
              >
                {{ $t('workflow.batchGenerateProps') }} ({{ selectedPropIds.length }})
              </el-button>
            </div>
          </div>
          
          <div class="scene-image-list">
            <div v-for="prop in propsForEpisode" :key="prop.id" class="scene-item">
              <el-card shadow="hover" class="fixed-card">
                <div class="card-header">
                  <el-checkbox 
                    v-model="selectedPropIds"
                    :value="prop.id"
                    style="margin-right: 8px;"
                  />
                  <div class="header-left">
                    <h4>{{ prop.name }}</h4>
                    <el-tag v-if="prop.type" size="small">{{ prop.type }}</el-tag>
                  </div>
                </div>

                <div class="card-image-container">
                  <div v-if="prop.image_url" class="prop-image">
                    <el-image 
                      :src="prop.image_url" 
                      fit="cover" 
                      :preview-src-list="[prop.image_url]"
                      :preview-teleported="true"
                    />
                  </div>
                  <div v-else-if="generatingPropImages[prop.id]" class="prop-placeholder generating">
                    <el-icon :size="64" class="rotating"><Loading /></el-icon>
                    <span>{{ $t('common.generating') }}</span>
                  </div>
                  <div v-else class="prop-placeholder">
                    <el-icon :size="64"><Box /></el-icon>
                    <span>{{ $t('common.notGenerated') }}</span>
                  </div>
                </div>

                <div class="card-actions">
                  <el-tooltip :content="$t('tooltip.editPrompt')" placement="top">
                    <el-button 
                      size="small" 
                      @click="openPromptDialog(prop, 'prop')"
                      :icon="Edit"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.aiGenerate')" placement="top">
                    <el-button 
                      type="primary"
                      size="small" 
                      @click="generatePropImage(prop)"
                      :loading="generatingPropImages[prop.id]"
                      :icon="MagicStick"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip :content="$t('tooltip.uploadImage')" placement="top">
                    <el-button 
                      size="small" 
                      @click="uploadPropImage(prop.id)"
                      :icon="Upload"
                      circle
                    />
                  </el-tooltip>
                  <el-tooltip content="引用已有素材" placement="top">
                    <el-button 
                      size="small" 
                      @click="openAssetReference('prop', prop)"
                      :icon="More"
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

        <el-form-item v-if="currentEditType === 'character' || currentEditType === 'prop'" label="辅助设置">
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
            <el-button size="small" :loading="optimizingPrompt" :disabled="optimizingPrompt" @click="optimizePrompt">一键优化提示词</el-button>
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

    <el-dialog 
      v-model="assetReferenceDialogVisible" 
      title="引用已有素材"
      width="900px"
    >
      <div style="display: flex; gap: 12px; margin-bottom: 16px;">
        <el-input 
          v-model="assetSearchQuery" 
          placeholder="搜索素材名称或描述" 
          clearable
          @keyup.enter="searchAssets"
        />
        <el-button type="primary" @click="searchAssets" :loading="assetReferenceLoading">
          搜索
        </el-button>
      </div>
      <el-table :data="assetReferenceItems" border stripe style="width: 100%;" v-loading="assetReferenceLoading">
        <el-table-column prop="name" label="素材名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column label="预览" width="120">
          <template #default="scope">
            <el-image 
              :src="scope.row.thumbnail_url || scope.row.url" 
              fit="cover" 
              style="width: 80px; height: 80px; border-radius: 6px;"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button type="primary" size="small" @click="associateAsset(scope.row)">
              引用
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="assetReferencePagination.total"
          :page-size="assetReferencePagination.page_size"
          :current-page="assetReferencePagination.page"
          @current-change="handleAssetPageChange"
        />
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
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { debounce } from 'lodash-es'
import { 
  User, 
  Location, 
  Picture,
  MagicStick,
  ArrowRight,
  ArrowLeft,
  Place,
  Box,
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
import { propAPI } from '@/api/prop'
import { assetAPI } from '@/api/asset'
import { aiAPI } from '@/api/ai'
import type { AIServiceConfig } from '@/types/ai'
import { imageAPI } from '@/api/image'
import type { Drama } from '@/types/drama'
import type { Prop } from '@/types/prop'
import type { Asset } from '@/types/asset'
import { AppHeader } from '@/components/common'

// 预设提示词配置
const PRESET_PROMPTS = {
  white_background: {
    inject: [
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
    ],
    detect: [
      'white background', 'simple background', 'solid background', 'blank background',
      '100% pure white solid background', 'pure white background', '纯白背景', '白底'
    ]
  },
  three_views: {
    inject: [
       'character turnaround sheet',
       'orthographic design drawing',
       'three orthogonal views of the same character on one single image',
       'front view', 'side view', 'back view',
       'no missing views',
       'consistent character details across all views',
       'professional character design sheet'
    ],
    detect: [
      'three views', 'front view', 'side view', 'back view', 'character turnaround', 'character sheet', 'orthographic',
      'three orthogonal views', '三视图', '正面', '侧面', '背面'
    ]
  },
  composition_fix: {
    inject: [
       'full body fully visible', 'whole body', 'no cropping', 'no out of frame',
       'no partial body', 'padding around edges', 'character centered',
       'clean and tidy layout', 'no ugly', 'no disfigured',
       'no malformed limbs', 'no extra limbs', 'no missing limbs'
    ],
    detect: [
      'character centered', 'full body fully visible', 'whole body', 'no cropping', 'padding around edges',
      'no out of frame', '中心构图', '对称', '黄金分割'
    ]
  }
}

const route = useRoute()
const router = useRouter()
const { t: $t } = useI18n()
const dramaId = route.params.id as string
const episodeNumber = parseInt(route.params.episodeNumber as string)

const drama = ref<Drama>()
const props = ref<Prop[]>([])

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
const generatingPropImages = ref<Record<number, boolean>>({})

// 选择状态
const selectedCharacterIds = ref<number[]>([])
const selectedSceneIds = ref<number[]>([])
const selectedPropIds = ref<number[]>([])
const selectAllCharacters = ref(false)
const selectAllScenes = ref(false)
const selectAllProps = ref(false)

// 对话框状态
const promptDialogVisible = ref(false)
const libraryDialogVisible = ref(false)
const uploadDialogVisible = ref(false)
const modelConfigDialogVisible = ref(false)
const assetReferenceDialogVisible = ref(false)
const currentEditItem = ref<any>({ name: '' })
const currentEditType = ref<'character' | 'scene' | 'prop'>('character')
const editPrompt = ref('')
const editStyle = ref('')
const editReference = ref('')
const editSize = ref('')
const editViewSettings = ref<string[]>([])
const previousViewSettings = ref<string[]>([])
// 一键优化提示词的撤销缓存
const optimizePromptHistory = ref('')
// 控制撤销按钮是否可用
const canUndoOptimize = ref(false)
const optimizingPrompt = ref(false)
const libraryItems = ref<any[]>([])
const assetSearchQuery = ref('')
const assetReferenceType = ref<'character' | 'scene' | 'prop'>('character')
const assetReferenceTarget = ref<any>(null)
const assetReferenceItems = ref<Asset[]>([])
const assetReferenceLoading = ref(false)
const assetReferencePagination = ref({
  page: 1,
  page_size: 20,
  total: 0
})
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

const propsForEpisode = computed(() => {
  return props.value
})

const propsCount = computed(() => {
  return propsForEpisode.value.length
})

const hasProps = computed(() => {
  return propsCount.value > 0
})

const hasExtractedData = computed(() => {
  const hasScenes = currentEpisode.value?.scenes && currentEpisode.value.scenes.length > 0
  // 只要有角色或场景，就认为已经提取过数据
  return hasCharacters.value || hasScenes || hasProps.value
})

const allImagesGenerated = computed(() => {
  if (!hasExtractedData.value) return true
  
  const characters = currentEpisode.value?.characters || []
  const scenes = currentEpisode.value?.scenes || []
  const currentProps = propsForEpisode.value || []
  
  if (characters.length === 0 && scenes.length === 0 && currentProps.length === 0) return true
  
  const allCharsHaveImages = characters.length === 0 || characters.every(char => char.image_url)
  const allScenesHaveImages = scenes.length === 0 || scenes.every(scene => scene.image_url)
  const allPropsHaveImages = currentProps.length === 0 || currentProps.every(prop => prop.image_url)
  
  return allCharsHaveImages && allScenesHaveImages && allPropsHaveImages
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

const loadProps = async () => {
  try {
    if (!currentEpisode.value?.id) {
      props.value = []
      return
    }
    const result = await propAPI.listByEpisode(currentEpisode.value.id)
    props.value = Array.isArray(result) ? result : []
  } catch (error: any) {
    console.error('加载道具失败:', error)
    props.value = []
  }
}

const loadDramaData = async () => {
  try {
    const data = await dramaAPI.get(dramaId)
    drama.value = data
    await loadProps()
    
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

  const scriptContent = currentEpisode.value.script_content || ''
  if (!scriptContent.trim()) {
    ElMessage.warning('剧本内容为空，请先生成或填写剧本')
    return
  }

  extractingCharactersAndBackgrounds.value = true
  
  try {
    const episodeId = currentEpisode.value.id

    // 并行创建异步任务
    const [characterTask, backgroundTask, propTask] = await Promise.all([
      generationAPI.generateCharacters({
        drama_id: dramaId.toString(),
        episode_id: parseInt(episodeId),
        outline: scriptContent,
        count: 0,
        model: selectedTextModel.value  // 传递用户选择的文本模型
      }),
      dramaAPI.extractBackgrounds(episodeId.toString(), selectedTextModel.value),
      propAPI.extractFromScript(parseInt(episodeId))
    ])
    
    ElMessage.success('任务已创建，正在后台处理...')
    
    // 并行轮询两个任务
    await Promise.all([
      pollExtractTask(characterTask.task_id, 'character'),
      pollExtractTask(backgroundTask.task_id, 'background'),
      pollExtractTask(propTask.task_id, 'prop')
    ])
    
    ElMessage.success('角色、场景和道具提取成功！')
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
    } else if (
      errorMsg.includes('script content') ||
      errorMsg.includes('剧本内容为空')
    ) {
      ElMessage.warning('剧本内容为空，请先生成或填写剧本')
    } else if (
      errorMsg.includes('episode not found') ||
      errorMsg.includes('剧集信息不存在') ||
      errorMsg.includes('章节信息不存在')
    ) {
      ElMessage.error('章节信息不存在，请刷新页面后重试')
    } else {
      ElMessage.error(errorMsg)
    }
  } finally {
    extractingCharactersAndBackgrounds.value = false
  }
}

// 轮询提取任务状态
const pollExtractTask = async (taskId: string, type: 'character' | 'background' | 'prop') => {
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
        const errorPrefix = type === 'character' ? '角色生成' : type === 'background' ? '场景提取' : '道具提取'
        const errorText = task.error || task.message || `${errorPrefix}失败`
        throw new Error(errorText)
      }
      // 否则继续轮询
    } catch (error: any) {
      console.error(`轮询${type}任务状态失败:`, error)
      throw error
    }
  }
  
  const timeoutPrefix = type === 'character' ? '角色生成' : type === 'background' ? '场景提取' : '道具提取'
  throw new Error(`${timeoutPrefix}超时`)
}


const generateCharacterImage = async (characterId: number, styleParam?: string, sizeParam?: string, referenceWorkParam?: string, whiteBackgroundParam?: boolean, forceRegenerate?: boolean) => {
  const currentCharacter = currentEpisode.value?.characters?.find(char => char.id === characterId)
  if (currentCharacter?.image_url && !forceRegenerate) {
    ElMessage.info('角色已有图片')
    return
  }

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
    ElMessage.error(formatImageGenerationError(error, '生成失败'))
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

const toggleSelectAllProps = () => {
  if (selectAllProps.value) {
    selectedPropIds.value = propsForEpisode.value.map(prop => prop.id)
  } else {
    selectedPropIds.value = []
  }
}

const formatImageGenerationError = (error: any, fallback: string) => {
  const message = error?.response?.data?.error?.message || error?.message || fallback
  const normalized = String(message || '').toLowerCase()
  if (
    normalized.includes('no such host') ||
    normalized.includes('dial tcp') ||
    normalized.includes('ark.cn-beijing.volces.com')
  ) {
    return 'AI服务地址无法解析，请在 设置 > AI服务配置 中切换可用服务或检查网络/DNS'
  }
  if (normalized.includes('no active config found')) {
    return '未找到可用的AI服务配置，请先在 设置 > AI服务配置 中启用配置'
  }
  return message || fallback
}

const batchGenerateCharacterImages = async () => {
  if (selectedCharacterIds.value.length === 0) {
    ElMessage.warning('请先选择要生成的角色')
    return
  }
  
  batchGeneratingCharacters.value = true
  try {
    const currentCharacters = currentEpisode.value?.characters || []
    const selectedCharacters = currentCharacters.filter(char => selectedCharacterIds.value.includes(char.id))
    const pendingCharacters = selectedCharacters.filter(char => !char.image_url)
    if (pendingCharacters.length === 0) {
      ElMessage.info('选中的角色都已有图片')
      return
    }

    // 获取用户选择的图片生成模型
    const model = selectedImageModel.value || undefined
    
    // 使用批量生成API
    await characterLibraryAPI.batchGenerateCharacterImages(
      pendingCharacters.map(char => char.id.toString()),
      model
    )
    
    ElMessage.success($t('workflow.batchTaskSubmitted'))
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(formatImageGenerationError(error, $t('workflow.batchGenerateFailed')))
  } finally {
    batchGeneratingCharacters.value = false
  }
}

const generateSceneImage = async (sceneId: number) => {
  const currentScene = currentEpisode.value?.scenes?.find(scene => scene.id === sceneId)
  if (currentScene?.image_url) {
    ElMessage.info('场景已有图片')
    return
  }

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
    ElMessage.error(formatImageGenerationError(error, '生成失败'))
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
    const currentScenes = currentEpisode.value?.scenes || []
    const selectedScenes = currentScenes.filter(scene => selectedSceneIds.value.includes(scene.id))
    const pendingScenes = selectedScenes.filter(scene => !scene.image_url)
    if (pendingScenes.length === 0) {
      ElMessage.info('选中的场景都已有图片')
      return
    }

    const promises = pendingScenes.map(scene => 
      generateSceneImage(scene.id)
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
    ElMessage.error(formatImageGenerationError(error, $t('workflow.batchGenerateFailed')))
  } finally {
    batchGeneratingScenes.value = false
  }
}

const generatePropImage = async (prop: any) => {
  if (!prop.prompt) {
    ElMessage.warning('请先设置道具的图片提示词')
    openPromptDialog(prop, 'prop')
    return
  }
  if (prop.image_url) {
    ElMessage.info('道具已有图片')
    return
  }

  generatingPropImages.value[prop.id] = true
  try {
    const response = await propAPI.generateImage(prop.id)
    const taskId = response.task_id
    if (taskId) {
      await pollSimpleTask(taskId, '道具图片生成完成！')
    } else {
      ElMessage.success('道具图片生成已启动')
      await loadDramaData()
    }
  } catch (error: any) {
    ElMessage.error(formatImageGenerationError(error, '生成失败'))
  } finally {
    generatingPropImages.value[prop.id] = false
  }
}

const batchGeneratePropImages = async () => {
  if (selectedPropIds.value.length === 0) {
    ElMessage.warning('请先选择要生成的道具')
    return
  }

  try {
    const selectedProps = propsForEpisode.value.filter(prop => selectedPropIds.value.includes(prop.id))
    const pendingProps = selectedProps.filter(prop => !prop.image_url)
    if (pendingProps.length === 0) {
      ElMessage.info('选中的道具都已有图片')
      return
    }
    await Promise.all(pendingProps.map(prop => propAPI.generateImage(prop.id)))
    ElMessage.success('道具批量生成任务已提交')
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || '批量生成失败')
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
    ElMessage.error(formatImageGenerationError(error, '拆分失败'))
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
        ElMessage.error(formatImageGenerationError({ message: task.error }, '分镜拆分失败'))
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
const openPromptDialog = (item: any, type: 'character' | 'scene' | 'prop') => {
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
  if (type === 'character' || type === 'prop') {
      const settings: string[] = []
      const currentPrompt = editPrompt.value.toLowerCase()
      
      // 白底图 (检测或为空时默认选中)
      const hasWhiteBg = PRESET_PROMPTS.white_background.detect.some(k => currentPrompt.includes(k.toLowerCase()))
      if (hasWhiteBg || !currentPrompt) {
          settings.push('white_background')
      }
      
      // 三视图 (检测或为空时默认选中)
      const hasThreeViews = PRESET_PROMPTS.three_views.detect.some(k => currentPrompt.includes(k.toLowerCase()))
      if (hasThreeViews || !currentPrompt) {
          settings.push('three_views')
      }
      
      // 构图修正 (检测或为空时默认选中)
      const hasComposition = PRESET_PROMPTS.composition_fix.detect.some(k => currentPrompt.includes(k.toLowerCase()))
      if (hasComposition || !currentPrompt) {
          settings.push('composition_fix')
      }
      
      editViewSettings.value = settings
      if (settings.length > 0) {
          updatePromptWithSettings(settings, [])
      }
  } else {
      editViewSettings.value = []
  }
  
  previousViewSettings.value = [...editViewSettings.value]
  promptDialogVisible.value = true
}

// 监听视图设置变化，实时更新提示词
const handleViewSettingsChange = (newSettings: string[]) => {
  const oldSettings = previousViewSettings.value
  previousViewSettings.value = [...newSettings]
  
  updatePromptWithSettings(newSettings, oldSettings)
}

// 监听视图设置变化（作为 v-model 的补充，确保实时性）
watch(editViewSettings, (newSettings) => {
  if (promptDialogVisible.value) {
    handleViewSettingsChange(newSettings)
  }
}, { deep: true })

const updatePromptWithSettings = (newSettings: string[], oldSettings: string[]) => {
  let current = editPrompt.value
  
  // 找出新增的选项
  const added = newSettings.filter(s => !oldSettings.includes(s))
  // 找出移除的选项
  const removed = oldSettings.filter(s => !newSettings.includes(s))
  
  // 处理移除
  if (removed.length > 0) {
    let segments = current.split(/[,，\n]+/).map(s => s.trim()).filter(Boolean)
    
    removed.forEach(setting => {
      const config = PRESET_PROMPTS[setting as keyof typeof PRESET_PROMPTS]
      if (config) {
        // 移除所有相关的提示词（包括检测到的和注入的）
        const allKeywords = [...config.detect, ...config.inject]
        segments = segments.filter(seg => {
          const lowerSeg = seg.toLowerCase()
          return !allKeywords.some(k => lowerSeg.includes(k.toLowerCase()))
        })
      }
    })
    current = segments.join(', ')
  }
  
  // 处理新增
  if (added.length > 0) {
    let segments = current.split(/[,，\n]+/).map(s => s.trim()).filter(Boolean)
    
    added.forEach(setting => {
       const config = PRESET_PROMPTS[setting as keyof typeof PRESET_PROMPTS]
       if (config) {
         const newKeywords = config.inject.filter(k => 
           !segments.some(seg => seg.toLowerCase().includes(k.toLowerCase()))
         )
         if (newKeywords.length > 0) {
           segments.push(...newKeywords)
         }
       }
    })
    current = segments.join(', ')
  }
  
  editPrompt.value = current
}

const getProtectedKeywords = () => {
  const keywords = new Set<string>()
  editViewSettings.value.forEach(setting => {
    const config = PRESET_PROMPTS[setting as keyof typeof PRESET_PROMPTS]
    if (config) {
      config.inject.forEach(item => keywords.add(item))
    }
  })
  if (editStyle.value) {
    keywords.add(editStyle.value)
  }
  if (editReference.value) {
    keywords.add(`style reference: ${editReference.value}`)
  }
  return Array.from(keywords)
}

// 保护性提示词识别与优化处理
const optimizePrompt = async () => {
  if (!editPrompt.value || !editPrompt.value.trim()) {
    ElMessage.warning('提示词为空，无法优化')
    return
  }
  if (optimizingPrompt.value) {
    return
  }
  optimizingPrompt.value = true
  optimizePromptHistory.value = editPrompt.value
  try {
    const protectedKeywords = getProtectedKeywords()
    const response = await aiAPI.optimizePrompt(editPrompt.value, protectedKeywords)
    const optimized = response.prompt?.trim()
    if (optimized) {
      editPrompt.value = optimized
      canUndoOptimize.value = true
      ElMessage.success('提示词已优化，已保护白底、三视图、基础风格及构图关键词')
      return
    }
    const fallback = buildOptimizedPrompt(editPrompt.value)
    editPrompt.value = fallback
    canUndoOptimize.value = true
    ElMessage.success('提示词已优化，已使用本地优化策略')
  } catch (error: any) {
    const fallback = buildOptimizedPrompt(editPrompt.value)
    editPrompt.value = fallback
    canUndoOptimize.value = true
    ElMessage.warning('提示词优化服务不可用，已使用本地优化策略')
  } finally {
    optimizingPrompt.value = false
  }
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

  // 白底图保护关键词 (包含检测和注入的词)
  const whiteBgProtected = [
    ...PRESET_PROMPTS.white_background.detect,
    ...PRESET_PROMPTS.white_background.inject
  ]
  // 三视图保护关键词
  const threeViewsProtected = [
    ...PRESET_PROMPTS.three_views.detect,
    ...PRESET_PROMPTS.three_views.inject
  ]
  // 构图修正保护关键词
  const compositionProtected = [
    ...PRESET_PROMPTS.composition_fix.detect,
    ...PRESET_PROMPTS.composition_fix.inject
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
    /cinematic|highly detailed|细节|电影质感|best quality|masterpiece/i.test(seg)
  )
  if (!hasStyleEnhancement && rawPrompt.length < 800) {
    resultSegments.push('best quality, masterpiece, highly detailed')
  }

  return resultSegments.join(', ')
}

const savePrompt = async () => {
  try {
    if (currentEditType.value === 'character') {
      // 提示词已实时同步，直接使用
      const finalPrompt = editPrompt.value
      
      await characterLibraryAPI.updateCharacter(currentEditItem.value.id, {
        appearance: finalPrompt
      })

      if (currentEditItem.value) {
        currentEditItem.value.appearance = finalPrompt
      }
      if (currentEpisode.value?.characters) {
        const targetCharacter = currentEpisode.value.characters.find(char => char.id === currentEditItem.value.id)
        if (targetCharacter) {
          targetCharacter.appearance = finalPrompt
        }
      }
      
      // 计算是否启用白底图
      const useWhiteBackground = editViewSettings.value.includes('white_background')
      
      // 传递自定义参数进行生成
      await generateCharacterImage(currentEditItem.value.id, editStyle.value, editSize.value, editReference.value, useWhiteBackground, true)
    } else if (currentEditType.value === 'prop') {
      const finalPrompt = editPrompt.value
      
      // 保存道具提示词
      await propAPI.update(currentEditItem.value.id, {
        prompt: finalPrompt
      })

      if (currentEditItem.value) {
        currentEditItem.value.prompt = finalPrompt
      }
      const targetProp = props.value.find(prop => prop.id === currentEditItem.value.id)
      if (targetProp) {
        targetProp.prompt = finalPrompt
      }
      
      // 启动道具图片生成并保持生成状态
      generatingPropImages.value[currentEditItem.value.id] = true
      try {
        const response = await propAPI.generateImage(currentEditItem.value.id)
        const taskId = response.task_id
        // 有任务ID时轮询等待结果
        if (taskId) {
          await pollSimpleTask(taskId, '道具图片生成完成！')
        } else {
          ElMessage.success('道具图片生成已启动')
          await loadDramaData()
        }
      } catch (error: any) {
        ElMessage.error(formatImageGenerationError(error, '生成失败'))
      } finally {
        generatingPropImages.value[currentEditItem.value.id] = false
      }
    } else {
      // 1. 先保存场景提示词
      await dramaAPI.updateScenePrompt(currentEditItem.value.id.toString(), editPrompt.value)

      if (currentEditItem.value) {
        currentEditItem.value.prompt = editPrompt.value
      }
      if (currentEpisode.value?.scenes) {
        const targetScene = currentEpisode.value.scenes.find(scene => scene.id === currentEditItem.value.id)
        if (targetScene) {
          targetScene.prompt = editPrompt.value
        }
      }
      
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

const uploadPropImage = (propId: number) => {
  currentUploadTarget.value = { id: propId, type: 'prop' }
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

const getAssetCategory = (type: 'character' | 'scene' | 'prop') => {
  if (type === 'character') return '角色'
  if (type === 'scene') return '场景'
  return '道具'
}

const openAssetReference = async (type: 'character' | 'scene' | 'prop', target: any) => {
  assetReferenceType.value = type
  assetReferenceTarget.value = target
  assetReferencePagination.value.page = 1
  assetReferenceDialogVisible.value = true
  await searchAssets()
}

const searchAssets = async () => {
  try {
    assetReferenceLoading.value = true
    const category = getAssetCategory(assetReferenceType.value)
    const result = await assetAPI.listAssets({
      drama_id: dramaId,
      include_shared: true,
      type: 'image',
      category,
      search: assetSearchQuery.value.trim() || undefined,
      page: assetReferencePagination.value.page,
      page_size: assetReferencePagination.value.page_size
    })
    assetReferenceItems.value = result.items || []
    assetReferencePagination.value.total = result.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '加载素材失败')
  } finally {
    assetReferenceLoading.value = false
  }
}

const debouncedSearchAssets = debounce(() => {
  if (!assetReferenceDialogVisible.value) {
    return
  }
  assetReferencePagination.value.page = 1
  searchAssets()
}, 400)

watch(assetSearchQuery, (value, prev) => {
  if (value === prev) {
    return
  }
  debouncedSearchAssets()
})

onBeforeUnmount(() => {
  debouncedSearchAssets.cancel()
})

const handleAssetPageChange = (page: number) => {
  assetReferencePagination.value.page = page
  searchAssets()
}

const associateAsset = async (asset: Asset) => {
  if (!assetReferenceTarget.value) {
    return
  }
  try {
    const imageUrl = asset.url
    if (assetReferenceType.value === 'character') {
      await characterLibraryAPI.updateCharacter(assetReferenceTarget.value.id, {
        image_url: imageUrl
      })
    } else if (assetReferenceType.value === 'scene') {
      await dramaAPI.updateScene(assetReferenceTarget.value.id.toString(), {
        image_url: imageUrl
      })
    } else {
      await propAPI.update(assetReferenceTarget.value.id, {
        image_url: imageUrl
      })
    }
    ElMessage.success('引用素材成功')
    assetReferenceDialogVisible.value = false
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || '引用素材失败')
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
    } else if (currentUploadTarget.value?.type === 'prop') {
      await propAPI.update(currentUploadTarget.value.id, { image_url: imageUrl })
      ElMessage.success('上传成功！')
    }
    
    await loadDramaData()
    uploadDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  }
}

const pollSimpleTask = async (taskId: string, successMessage: string) => {
  const maxAttempts = 300
  const interval = 2000

  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(resolve => setTimeout(resolve, interval))

    const task = await generationAPI.getTaskStatus(taskId)
    if (task.status === 'completed') {
      await loadDramaData()
      ElMessage.success(successMessage)
      return
    }
    if (task.status === 'failed') {
      throw new Error(task.error || '任务失败')
    }
  }

  throw new Error('任务超时')
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
    .scene-image,
    .prop-image {
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
    .scene-placeholder,
    .prop-placeholder {
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
