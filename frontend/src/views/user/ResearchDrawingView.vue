<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-6xl space-y-6">
      <div class="card overflow-hidden">
        <div class="p-4 sm:p-5">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div class="max-w-3xl">
              <h3 class="text-xl font-bold text-gray-900 dark:text-white">
                {{ t('researchDrawing.topNote') }}
              </h3>
              <p class="mt-2 max-w-3xl text-sm leading-5 text-gray-600 dark:text-dark-300">
                {{ t('researchDrawing.description') }}
              </p>
            </div>
            <p class="shrink-0 rounded-md bg-primary-50 px-3 py-1.5 text-sm font-semibold text-primary-700 dark:bg-primary-950/30 dark:text-primary-300">
              {{ t('researchDrawing.unitPriceNote') }}
            </p>
          </div>

          <div class="mt-4 border-t border-gray-100 pt-3 dark:border-dark-700">
            <div class="flex flex-col gap-2 lg:flex-row lg:items-start">
              <h4 class="shrink-0 text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.tipsTitle') }}</h4>
              <div class="grid flex-1 gap-2 text-xs leading-5 text-gray-600 dark:text-dark-300 md:grid-cols-3">
                <p>{{ t('researchDrawing.tips.first') }}</p>
                <p>{{ t('researchDrawing.tips.second') }}</p>
                <p>{{ t('researchDrawing.tips.third') }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <section class="space-y-4">
        <div>
          <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.examplesTitle') }}</h4>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('researchDrawing.examplesDesc') }}</p>
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <article
            v-for="item in exampleCards"
            :key="item.title"
            class="card overflow-hidden"
          >
            <div class="aspect-[4/3] bg-gray-50 dark:bg-dark-900">
              <img
                class="h-full w-full object-contain p-3"
                :src="item.image"
                :alt="item.title"
                loading="lazy"
                decoding="async"
              />
            </div>
            <div class="p-4">
              <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</h5>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ item.desc }}</p>
            </div>
          </article>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-6 lg:grid-cols-10 lg:items-start">
        <form class="card self-start space-y-5 p-6 lg:col-span-6" @submit.prevent="startGenerationPreview">
          <div class="border-b border-gray-100 pb-4 dark:border-dark-700">
            <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.input.title') }}</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('researchDrawing.input.desc') }}</p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.input.loadMethodExample') }}</span>
              <select v-model="generationInput.methodExample" class="input" @change="loadMethodExample">
                <option value="">{{ t('researchDrawing.input.noExample') }}</option>
                <option value="paperVizAgent">{{ t('researchDrawing.input.examples.paperVizAgentFramework') }}</option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.input.loadCaptionExample') }}</span>
              <select v-model="generationInput.captionExample" class="input" @change="loadCaptionExample">
                <option value="">{{ t('researchDrawing.input.noExample') }}</option>
                <option value="paperVizAgent">{{ t('researchDrawing.input.examples.paperVizAgentFramework') }}</option>
              </select>
            </label>
          </div>

          <label class="field-wrap">
            <span>{{ t('researchDrawing.input.methodContent') }}</span>
            <textarea
              v-model="generationInput.methodContent"
              class="input min-h-[220px] resize-y"
              required
              :placeholder="t('researchDrawing.input.placeholders.methodContent')"
            ></textarea>
          </label>

          <div v-if="isCustomGenerationMode" class="rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
            <label class="flex items-start gap-3 text-sm text-gray-700 dark:text-dark-300">
              <input
                v-model="generationInput.optimizeMethodContent"
                class="mt-1 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                type="checkbox"
                :disabled="!methodOptimizationEnabled"
              />
              <span>{{ t('researchDrawing.input.methodOptimize') }}</span>
            </label>
            <p v-if="!methodOptimizationEnabled" class="mt-2 text-sm text-gray-500 dark:text-dark-400">
              {{ t('researchDrawing.input.optimizeUnavailable') }}
            </p>
          </div>

          <label class="field-wrap">
            <span>{{ t('researchDrawing.input.caption') }}</span>
            <textarea
              v-model="generationInput.caption"
              class="input min-h-[112px] resize-y"
              :placeholder="t('researchDrawing.input.placeholders.caption')"
            ></textarea>
          </label>

          <template v-if="isAdmin">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.input.generationMode') }}</span>
              <select v-model="generationInput.generationMode" class="input">
                <option value="default">{{ t('researchDrawing.input.defaultMode') }}</option>
                <option value="custom">{{ t('researchDrawing.input.customMode') }}</option>
              </select>
            </label>

            <p
              v-if="generationInput.generationMode === 'custom'"
              class="rounded-lg bg-primary-50 p-3 text-sm text-primary-700 dark:bg-primary-950/30 dark:text-primary-300"
            >
              {{ t('researchDrawing.input.customModeHint') }}
            </p>

            <div v-if="generationInput.generationMode === 'custom'" class="space-y-4">
              <div class="grid gap-4 md:grid-cols-3">
                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.expMode') }}</span>
                  <select v-model="form.research_drawing_exp_mode" class="input">
                    <option value="demo_planner_critic">规划与评审</option>
                    <option value="demo_full">完整流程</option>
                  </select>
                </label>

                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.retrievalSetting') }}</span>
                  <select v-model="form.research_drawing_retrieval_setting" class="input">
                    <option value="auto">自动</option>
                    <option value="manual">手动</option>
                    <option value="random">随机</option>
                    <option value="none">不检索</option>
                  </select>
                </label>

                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.numCandidates') }}</span>
                  <input v-model.number="form.research_drawing_num_candidates" class="input" type="number" min="1" max="20" />
                </label>
              </div>

              <div class="grid gap-4 md:grid-cols-3">
                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.aspectRatio') }}</span>
                  <select v-model="form.research_drawing_aspect_ratio" class="input">
                    <option value="16:9">16:9</option>
                    <option value="21:9">21:9</option>
                    <option value="3:2">3:2</option>
                  </select>
                </label>

                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.maxCriticRounds') }}</span>
                  <input v-model.number="form.research_drawing_max_critic_rounds" class="input" type="number" min="1" max="5" />
                </label>

                <label class="field-wrap">
                  <span>{{ t('researchDrawing.labels.mainModelName') }}</span>
                  <select v-model="form.research_drawing_main_model_name" class="input">
                    <option
                      v-for="option in mainModelOptions"
                      :key="option.value"
                      :value="option.value"
                    >
                      {{ option.label }}
                    </option>
                  </select>
                </label>
              </div>

              <label class="field-wrap">
                <span>{{ t('researchDrawing.labels.imageGenModelName') }}</span>
                <select v-model="form.research_drawing_image_gen_model_name" class="input">
                  <option
                    v-for="option in imageModelOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </label>
            </div>
          </template>

          <p v-else class="rounded-lg bg-gray-50 p-3 text-sm text-gray-500 dark:bg-dark-900 dark:text-dark-400">
            {{ t('researchDrawing.input.normalUserHint') }}
          </p>

          <div class="rounded-lg border border-gray-100 bg-gray-50 p-4 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300">
            <span>{{ t('researchDrawing.run.estimatedTime') }}：</span>
            <b class="text-gray-900 dark:text-white">{{ t('researchDrawing.run.noHistory') }}</b>
          </div>

          <div class="flex flex-wrap gap-3">
            <button class="btn btn-primary" type="submit" :disabled="runSubmitting">
              {{ runSubmitting ? t('common.processing') : t('researchDrawing.run.start') }}
            </button>
            <button class="btn btn-secondary" type="button" @click="appStore.showInfo(t('researchDrawing.run.applyQuotaHint'))">
              {{ t('researchDrawing.run.applyQuota') }}
            </button>
            <button class="btn btn-secondary" type="button" @click="resetGenerationInput">
              {{ t('common.reset') }}
            </button>
          </div>

          <p
            v-if="runPreviewStarted"
            class="rounded-lg border border-primary-200 bg-primary-50 p-3 text-sm text-primary-700 dark:border-primary-900 dark:bg-primary-950/30 dark:text-primary-300"
          >
            {{ runStatusText }}
          </p>
        </form>

        <aside class="card self-start space-y-5 p-6 lg:col-span-4">
          <div class="border-b border-gray-100 pb-4 dark:border-dark-700">
            <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.title') }}</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.desc') }}</p>
          </div>

          <dl class="grid grid-cols-2 gap-3 text-sm">
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.unitPrice') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">2.99/次</dd>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.quotaNeed') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ quotaNeed }}</dd>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.input.generationMode') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">
                {{ isCustomGenerationMode ? t('researchDrawing.input.customMode') : t('researchDrawing.input.defaultMode') }}
              </dd>
            </div>
            <div v-if="isCustomGenerationMode" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.labels.aspectRatio') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ form.research_drawing_aspect_ratio }}</dd>
            </div>
            <div v-if="isCustomGenerationMode" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.labels.numCandidates') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ form.research_drawing_num_candidates }}</dd>
            </div>
            <div v-if="isCustomGenerationMode" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.labels.maxCriticRounds') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ form.research_drawing_max_critic_rounds }}</dd>
            </div>
            <div v-if="isCustomGenerationMode" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.labels.maxRefineResolution') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ form.research_drawing_max_refine_resolution }}</dd>
            </div>
          </dl>

          <div>
            <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.routesTitle') }}</h5>
            <div class="mt-3 space-y-2 text-sm text-gray-600 dark:text-dark-300">
              <p>{{ t('researchDrawing.run.routes.submit') }}</p>
              <p>{{ t('researchDrawing.run.routes.wait') }}</p>
              <p>{{ t('researchDrawing.run.routes.poll') }}</p>
              <p>{{ t('researchDrawing.run.routes.finish') }}</p>
            </div>
          </div>

          <div>
            <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.stepsTitle') }}</h5>
            <ol class="mt-3 space-y-3">
              <li
                v-for="(step, index) in runSteps"
                :key="step"
                class="flex gap-3 text-sm text-gray-600 dark:text-dark-300"
              >
                <span
                  class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
                  :class="runPreviewStarted || index === 0
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300'
                    : 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-dark-400'"
                >
                  {{ index + 1 }}
                </span>
                <span class="pt-0.5">{{ step }}</span>
              </li>
            </ol>
          </div>

          <div>
            <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.statusTitle') }}</h5>
            <div class="mt-3 space-y-2 text-sm text-gray-600 dark:text-dark-300">
              <p>{{ t('researchDrawing.run.statuses.queued') }}</p>
              <p>{{ t('researchDrawing.run.statuses.running', { time: '1 分 20 秒' }) }}</p>
              <p>{{ t('researchDrawing.run.statuses.done') }}</p>
              <p>{{ t('researchDrawing.run.statuses.error') }}</p>
              <p>{{ t('researchDrawing.run.statuses.missing') }}</p>
            </div>
          </div>

          <p
            class="rounded-lg border border-dashed p-3 text-sm leading-6"
            :class="runPreviewStarted
              ? 'border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-900 dark:bg-primary-950/30 dark:text-primary-300'
              : 'border-gray-200 bg-gray-50 text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400'"
          >
            {{ runPreviewStarted ? t('researchDrawing.run.previewStatus') : t('researchDrawing.run.idleStatus') }}
          </p>
        </aside>
      </section>

      <div v-if="loading" class="card flex items-center justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <form v-else-if="isAdmin" class="card space-y-6 p-6" @submit.prevent="saveSettings">
        <div class="border-b border-gray-100 pb-4 dark:border-dark-700">
          <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.settingsTitle') }}</h4>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('researchDrawing.settingsDesc') }}</p>
        </div>

        <section class="space-y-4">
          <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.sections.pipeline') }}</h4>
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.expMode') }}</span>
              <select v-model="form.research_drawing_exp_mode" class="input">
                <option value="demo_planner_critic">{{ t('researchDrawing.expMode.demoPlannerCritic') }}</option>
                <option value="demo_full">{{ t('researchDrawing.expMode.demoFull') }}</option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.retrievalSetting') }}</span>
              <select v-model="form.research_drawing_retrieval_setting" class="input">
                <option value="auto">{{ t('researchDrawing.retrieval.auto') }}</option>
                <option value="manual">{{ t('researchDrawing.retrieval.manual') }}</option>
                <option value="random">{{ t('researchDrawing.retrieval.random') }}</option>
                <option value="none">{{ t('researchDrawing.retrieval.none') }}</option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.aspectRatio') }}</span>
              <select v-model="form.research_drawing_aspect_ratio" class="input">
                <option value="16:9">16:9</option>
                <option value="21:9">21:9</option>
                <option value="3:2">3:2</option>
              </select>
            </label>
          </div>
        </section>

        <section class="space-y-4">
          <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.sections.generation') }}</h4>
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.unitPrice') }}</span>
              <input
                v-model.number="form.research_drawing_unit_price"
                class="input"
                type="number"
                min="0.01"
                step="0.01"
              />
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.numCandidates') }}</span>
              <input
                v-model.number="form.research_drawing_num_candidates"
                class="input"
                type="number"
                min="1"
                max="20"
              />
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.maxCriticRounds') }}</span>
              <input
                v-model.number="form.research_drawing_max_critic_rounds"
                class="input"
                type="number"
                min="1"
                max="5"
              />
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.maxRefineResolution') }}</span>
              <select v-model="form.research_drawing_max_refine_resolution" class="input">
                <option value="2K">2K</option>
                <option value="4K">4K</option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.methodOptimizationEnabled') }}</span>
              <select v-model="form.research_drawing_method_optimization_enabled" class="input">
                <option :value="true">{{ t('researchDrawing.options.enabled') }}</option>
                <option :value="false">{{ t('researchDrawing.options.disabled') }}</option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.methodOptimizationDefaultEnabled') }}</span>
              <select v-model="form.research_drawing_method_optimization_default_enabled" class="input">
                <option :value="true">{{ t('researchDrawing.options.enabled') }}</option>
                <option :value="false">{{ t('researchDrawing.options.disabled') }}</option>
              </select>
            </label>
          </div>
        </section>

        <section class="space-y-4">
          <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.sections.models') }}</h4>
          <div class="grid gap-4 md:grid-cols-2">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.mainModelName') }}</span>
              <select v-model="form.research_drawing_main_model_name" class="input">
                <option
                  v-for="option in mainModelOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.imageGenModelName') }}</span>
              <select v-model="form.research_drawing_image_gen_model_name" class="input">
                <option
                  v-for="option in imageModelOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>
          </div>
        </section>

        <div class="flex flex-wrap gap-3">
          <button class="btn btn-primary" type="submit" :disabled="saving">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
          <button class="btn btn-secondary" type="button" :disabled="saving" @click="resetToDefaults">
            {{ t('researchDrawing.resetDefaults') }}
          </button>
        </div>
      </form>

    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { SystemSettings, UpdateSettingsRequest } from '@/api/admin/settings'
import { researchDrawingAPI } from '@/api/researchDrawing'
import type { ResearchDrawingGenerateRequest, ResearchDrawingJobStatus } from '@/api/researchDrawing'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'

type ResearchDrawingForm = {
  research_drawing_exp_mode: string
  research_drawing_retrieval_setting: string
  research_drawing_num_candidates: number
  research_drawing_aspect_ratio: string
  research_drawing_max_critic_rounds: number
  research_drawing_main_model_name: string
  research_drawing_image_gen_model_name: string
  research_drawing_max_refine_resolution: string
  research_drawing_unit_price: number
  research_drawing_method_optimization_enabled: boolean
  research_drawing_method_optimization_default_enabled: boolean
}

type PaperBananaGenerationInput = {
  methodExample: string
  captionExample: string
  methodContent: string
  optimizeMethodContent: boolean
  caption: string
  generationMode: string
}

type RunResultImage = {
  candidateId: number
  url: string
}

const DEFAULT_EXAMPLE_METHOD = `## 方法：PaperVizAgent 框架

本节介绍 PaperVizAgent 的整体架构。PaperVizAgent 是一个参考驱动的智能体框架，用于自动生成学术插图。如图 \\ref{fig:methodology_diagram} 所示，PaperVizAgent 协调五类专门智能体：检索、规划、风格、可视化和评审，将原始科研内容转换为可用于论文发表的图示与图表。

### 检索智能体

给定源文本上下文 $S$ 和表达意图 $C$，检索智能体会从固定参考集中找到最相关的示例，用于指导后续生成过程。

### 规划智能体

规划智能体将源文本上下文和检索到的参考示例转换为目标插图的完整文字方案。

### 风格智能体

风格智能体根据学术审美进一步优化方案，包括配色、布局、字体和整体视觉一致性。

### 可视化与评审循环

可视化智能体根据优化后的方案生成候选图，评审智能体检查事实一致性和视觉质量，并提出改进提示词。该循环会进行多轮迭代，以获得接近发表质量的科研图。`

const DEFAULT_EXAMPLE_CAPTION =
  '图 1：PaperVizAgent 框架概览。给定源文本上下文和表达意图后，系统首先检索相关参考示例，并合成经过风格优化的描述。随后通过可视化与评审循环进行多轮细化，最终生成学术图。'

const GPT_IMAGE_2_MODEL = 'gpt-image-2'
const GPT_5_5_MODEL = 'gpt-5.5'

const mainModelOptions = [
  {
    label: 'Gemini 3 Flash Preview',
    value: 'openrouter/google/gemini-3-flash-preview',
  },
  {
    label: 'GPT-5.5',
    value: GPT_5_5_MODEL,
  },
]

const allowedMainModelValues = new Set(mainModelOptions.map((option) => option.value))

const imageModelOptions = [
  {
    label: 'Gemini 3.1 Flash Image Preview',
    value: 'openrouter/google/gemini-3.1-flash-image-preview',
  },
  {
    label: 'GPT-5.5-Image-2',
    value: GPT_IMAGE_2_MODEL,
  },
]

const allowedImageModelValues = new Set(imageModelOptions.map((option) => option.value))

const RESEARCH_DRAWING_DEFAULTS: ResearchDrawingForm = {
  research_drawing_exp_mode: 'demo_planner_critic',
  research_drawing_retrieval_setting: 'auto',
  research_drawing_num_candidates: 1,
  research_drawing_aspect_ratio: '16:9',
  research_drawing_max_critic_rounds: 2,
  research_drawing_main_model_name: 'openrouter/google/gemini-3-flash-preview',
  research_drawing_image_gen_model_name: 'openrouter/google/gemini-3.1-flash-image-preview',
  research_drawing_max_refine_resolution: '2K',
  research_drawing_unit_price: 2.99,
  research_drawing_method_optimization_enabled: true,
  research_drawing_method_optimization_default_enabled: false,
}

const PAPERBANANA_INPUT_DEFAULTS: PaperBananaGenerationInput = {
  methodExample: '',
  captionExample: '',
  methodContent: '',
  optimizeMethodContent: false,
  caption: '',
  generationMode: 'default',
}

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(true)
const saving = ref(false)
const lastSettings = ref<SystemSettings | null>(null)
const runPreviewStarted = ref(false)
const runSubmitting = ref(false)
const runJobId = ref('')
const runPaperBananaUser = ref('')
const runJobStatus = ref<ResearchDrawingJobStatus | null>(null)
const runImageLoading = ref(false)
const runResultImages = ref<RunResultImage[]>([])
const selectedResultImage = ref<RunResultImage | null>(null)
let runPollTimer: number | null = null
const isAdmin = computed(() => authStore.isAdmin)

const form = reactive<ResearchDrawingForm>({
  ...RESEARCH_DRAWING_DEFAULTS,
})

const generationInput = reactive<PaperBananaGenerationInput>({
  ...PAPERBANANA_INPUT_DEFAULTS,
})

const isCustomGenerationMode = computed(() => generationInput.generationMode === 'custom')

const methodOptimizationEnabled = computed(() => form.research_drawing_method_optimization_enabled !== false)

const quotaNeed = computed(() => {
  const candidates = Math.max(1, Number(form.research_drawing_num_candidates) || 1)
  const criticRounds = Math.max(1, Number(form.research_drawing_max_critic_rounds) || 1)
  return candidates * (1 + criticRounds)
})

const runSteps = computed(() => [
  t('researchDrawing.run.steps.validate'),
  t('researchDrawing.run.steps.spawn'),
  t('researchDrawing.run.steps.poll'),
  t('researchDrawing.run.steps.finish'),
  t('researchDrawing.run.steps.export'),
])

const runStatusText = computed(() => {
  if (!runJobId.value) {
    return t('researchDrawing.run.runningBanner')
  }
  const status = runJobStatus.value?.status || 'running'
  if (status === 'done') {
    return t('researchDrawing.run.statuses.done')
  }
  if (status === 'error') {
    return runJobStatus.value?.error || t('researchDrawing.run.statuses.error')
  }
  const elapsed = runJobStatus.value?.elapsed_sec
  if (typeof elapsed === 'number') {
    return t('researchDrawing.run.statuses.running', { time: formatDuration(elapsed) })
  }
  return t('researchDrawing.run.statuses.queued')
})

const exampleCards = computed(() => [
  {
    title: t('researchDrawing.examples.frameworkTitle'),
    desc: t('researchDrawing.examples.frameworkDesc'),
    image: '/research-drawing/scenario_architecture1.png',
  },
  {
    title: t('researchDrawing.examples.workflowTitle'),
    desc: t('researchDrawing.examples.workflowDesc'),
    image: '/research-drawing/scenario_ablation.svg',
  },
  {
    title: t('researchDrawing.examples.datasetTitle'),
    desc: t('researchDrawing.examples.datasetDesc'),
    image: '/research-drawing/dataset-on-hf-xl.png',
  },
])

function loadMethodExample() {
  if (generationInput.methodExample === 'paperVizAgent') {
    generationInput.methodContent = DEFAULT_EXAMPLE_METHOD
  }
}

function loadCaptionExample() {
  if (generationInput.captionExample === 'paperVizAgent') {
    generationInput.caption = DEFAULT_EXAMPLE_CAPTION
  }
}

function normalizeFormValues() {
  if (form.research_drawing_exp_mode !== 'demo_planner_critic' && form.research_drawing_exp_mode !== 'demo_full') {
    form.research_drawing_exp_mode = RESEARCH_DRAWING_DEFAULTS.research_drawing_exp_mode
  }

  if (!['auto', 'manual', 'random', 'none'].includes(form.research_drawing_retrieval_setting)) {
    form.research_drawing_retrieval_setting = RESEARCH_DRAWING_DEFAULTS.research_drawing_retrieval_setting
  }

  if (!['16:9', '21:9', '3:2'].includes(form.research_drawing_aspect_ratio)) {
    form.research_drawing_aspect_ratio = RESEARCH_DRAWING_DEFAULTS.research_drawing_aspect_ratio
  }

  if (!['2K', '4K'].includes(form.research_drawing_max_refine_resolution)) {
    form.research_drawing_max_refine_resolution = RESEARCH_DRAWING_DEFAULTS.research_drawing_max_refine_resolution
  }

  form.research_drawing_num_candidates = Math.min(
    20,
    Math.max(1, Number(form.research_drawing_num_candidates) || RESEARCH_DRAWING_DEFAULTS.research_drawing_num_candidates),
  )

  form.research_drawing_max_critic_rounds = Math.min(
    5,
    Math.max(1, Number(form.research_drawing_max_critic_rounds) || RESEARCH_DRAWING_DEFAULTS.research_drawing_max_critic_rounds),
  )
  form.research_drawing_unit_price = Math.max(
    0.01,
    Number(form.research_drawing_unit_price) || RESEARCH_DRAWING_DEFAULTS.research_drawing_unit_price,
  )

  form.research_drawing_main_model_name =
    form.research_drawing_main_model_name?.trim() || RESEARCH_DRAWING_DEFAULTS.research_drawing_main_model_name
  if (!allowedMainModelValues.has(form.research_drawing_main_model_name)) {
    form.research_drawing_main_model_name = RESEARCH_DRAWING_DEFAULTS.research_drawing_main_model_name
  }
  form.research_drawing_image_gen_model_name =
    form.research_drawing_image_gen_model_name?.trim() || RESEARCH_DRAWING_DEFAULTS.research_drawing_image_gen_model_name
  if (!allowedImageModelValues.has(form.research_drawing_image_gen_model_name)) {
    form.research_drawing_image_gen_model_name = RESEARCH_DRAWING_DEFAULTS.research_drawing_image_gen_model_name
  }
  if (!form.research_drawing_method_optimization_enabled) {
    form.research_drawing_method_optimization_default_enabled = false
  }
}

function applySettings(settings: SystemSettings) {
  lastSettings.value = settings
  form.research_drawing_exp_mode = settings.research_drawing_exp_mode || RESEARCH_DRAWING_DEFAULTS.research_drawing_exp_mode
  form.research_drawing_retrieval_setting = settings.research_drawing_retrieval_setting || RESEARCH_DRAWING_DEFAULTS.research_drawing_retrieval_setting
  form.research_drawing_num_candidates = settings.research_drawing_num_candidates || RESEARCH_DRAWING_DEFAULTS.research_drawing_num_candidates
  form.research_drawing_aspect_ratio = settings.research_drawing_aspect_ratio || RESEARCH_DRAWING_DEFAULTS.research_drawing_aspect_ratio
  form.research_drawing_max_critic_rounds = settings.research_drawing_max_critic_rounds || RESEARCH_DRAWING_DEFAULTS.research_drawing_max_critic_rounds
  form.research_drawing_main_model_name = settings.research_drawing_main_model_name || RESEARCH_DRAWING_DEFAULTS.research_drawing_main_model_name
  form.research_drawing_image_gen_model_name = settings.research_drawing_image_gen_model_name || RESEARCH_DRAWING_DEFAULTS.research_drawing_image_gen_model_name
  form.research_drawing_max_refine_resolution = settings.research_drawing_max_refine_resolution || RESEARCH_DRAWING_DEFAULTS.research_drawing_max_refine_resolution
  form.research_drawing_unit_price = settings.research_drawing_unit_price || RESEARCH_DRAWING_DEFAULTS.research_drawing_unit_price
  form.research_drawing_method_optimization_enabled = settings.research_drawing_method_optimization_enabled !== false
  form.research_drawing_method_optimization_default_enabled = settings.research_drawing_method_optimization_default_enabled === true
  applyMethodOptimizationDefault()
  normalizeFormValues()
}

function resetToDefaults() {
  Object.assign(form, RESEARCH_DRAWING_DEFAULTS)
}

function resetGenerationInput() {
  Object.assign(generationInput, PAPERBANANA_INPUT_DEFAULTS)
  applyMethodOptimizationDefault()
  runPreviewStarted.value = false
  runJobId.value = ''
  runPaperBananaUser.value = ''
  runJobStatus.value = null
  runImageLoading.value = false
  selectedResultImage.value = null
  revokeRunImages()
  stopRunPolling()
}

async function startGenerationPreview() {
  if (!generationInput.methodContent.trim()) {
    appStore.showWarning(t('researchDrawing.input.validationRequired'))
    return
  }
  if (!methodOptimizationEnabled.value) {
    generationInput.optimizeMethodContent = false
  }

  runSubmitting.value = true
  runPreviewStarted.value = true
  runJobStatus.value = { status: 'running' }
  selectedResultImage.value = null
  revokeRunImages()
  stopRunPolling()
  try {
    const payload: ResearchDrawingGenerateRequest = {
      method_content: generationInput.methodContent,
      caption: generationInput.caption,
      optimize_method_content: isCustomGenerationMode.value ? generationInput.optimizeMethodContent : false,
      generation_mode: generationInput.generationMode,
      ...(isCustomGenerationMode.value
        ? {
            exp_mode: form.research_drawing_exp_mode,
            retrieval_setting: form.research_drawing_retrieval_setting,
            num_candidates: form.research_drawing_num_candidates,
            aspect_ratio: form.research_drawing_aspect_ratio,
            max_critic_rounds: form.research_drawing_max_critic_rounds,
            main_model_name: form.research_drawing_main_model_name,
            image_gen_model_name: form.research_drawing_image_gen_model_name,
          }
        : {}),
    }
    const result = await researchDrawingAPI.generate(payload)
    runJobId.value = result.job_id
    runPaperBananaUser.value = result.paperbanana_user || ''
    appStore.showInfo(t('researchDrawing.run.submittedWithCharge', { charge: result.charge }))
    startRunPolling()
  } catch (error) {
    runPreviewStarted.value = false
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.run.submitFailed')))
  } finally {
    runSubmitting.value = false
  }
}

function startRunPolling() {
  if (!runJobId.value) {
    return
  }
  pollRunStatus()
  runPollTimer = window.setInterval(pollRunStatus, 2000)
}

function stopRunPolling() {
  if (runPollTimer) {
    window.clearInterval(runPollTimer)
    runPollTimer = null
  }
}

async function pollRunStatus() {
  if (!runJobId.value) {
    return
  }
  try {
    const status = await researchDrawingAPI.getJobStatus(runJobId.value, runPaperBananaUser.value)
    runJobStatus.value = status
    if (status.status === 'done' || status.status === 'error') {
      stopRunPolling()
      if (status.status === 'done') {
        await loadRunImages(status)
        appStore.showSuccess(t('researchDrawing.run.doneNotice'))
      } else {
        appStore.showError(status.error || t('researchDrawing.run.statuses.error'))
      }
    }
  } catch (error) {
    stopRunPolling()
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.run.statusFailed')))
  }
}

function getCandidateIds(status: ResearchDrawingJobStatus) {
  if (Array.isArray(status.images) && status.images.length > 0) {
    return status.images
      .map((image) => Number(image.candidate_id))
      .filter((candidateId) => Number.isInteger(candidateId) && candidateId >= 0)
  }
  if (Array.isArray(status.candidate_ids) && status.candidate_ids.length > 0) {
    return status.candidate_ids
      .map((candidateId) => Number(candidateId))
      .filter((candidateId) => Number.isInteger(candidateId) && candidateId >= 0)
  }
  const count = Math.max(0, Number(status.candidate_count) || 0)
  return Array.from({ length: count }, (_, index) => index)
}

async function loadRunImages(status: ResearchDrawingJobStatus) {
  const candidateIds = getCandidateIds(status)
  if (!runJobId.value || candidateIds.length === 0) {
    return
  }
  runImageLoading.value = true
  revokeRunImages()
  try {
    const images: RunResultImage[] = []
    for (const candidateId of candidateIds) {
      const blob = await researchDrawingAPI.getJobImage(runJobId.value, candidateId, runPaperBananaUser.value)
      images.push({
        candidateId,
        url: URL.createObjectURL(blob),
      })
    }
    runResultImages.value = images
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.run.imagesFailed')))
  } finally {
    runImageLoading.value = false
  }
}

function revokeRunImages() {
  runResultImages.value.forEach((image) => URL.revokeObjectURL(image.url))
  runResultImages.value = []
}

function formatDuration(seconds: number) {
  const sec = Math.max(0, Math.floor(seconds))
  const min = Math.floor(sec / 60)
  const rest = sec % 60
  return min > 0 ? `${min} 分 ${rest} 秒` : `${rest} 秒`
}

async function loadSettings() {
  if (!isAdmin.value) {
    loading.value = false
    return
  }

  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    applySettings(settings)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  if (!isAdmin.value || !lastSettings.value || saving.value) {
    return
  }

  normalizeFormValues()
  saving.value = true
  try {
    const payload: UpdateSettingsRequest = {
      ...lastSettings.value,
      research_drawing_exp_mode: form.research_drawing_exp_mode,
      research_drawing_retrieval_setting: form.research_drawing_retrieval_setting,
      research_drawing_num_candidates: form.research_drawing_num_candidates,
      research_drawing_aspect_ratio: form.research_drawing_aspect_ratio,
      research_drawing_max_critic_rounds: form.research_drawing_max_critic_rounds,
      research_drawing_main_model_name: form.research_drawing_main_model_name,
      research_drawing_image_gen_model_name: form.research_drawing_image_gen_model_name,
      research_drawing_max_refine_resolution: form.research_drawing_max_refine_resolution,
      research_drawing_unit_price: form.research_drawing_unit_price,
      research_drawing_method_optimization_enabled: form.research_drawing_method_optimization_enabled,
      research_drawing_method_optimization_default_enabled: form.research_drawing_method_optimization_default_enabled,
    }

    const updated = await adminAPI.settings.updateSettings(payload)
    applySettings(updated)
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.saveFailed')))
  } finally {
    saving.value = false
  }
}

function applyMethodOptimizationDefault() {
  generationInput.optimizeMethodContent =
    methodOptimizationEnabled.value && form.research_drawing_method_optimization_default_enabled
}

onMounted(() => {
  loadSettings()
})

onBeforeUnmount(() => {
  stopRunPolling()
  revokeRunImages()
})
</script>

<style scoped>
.field-wrap {
  @apply flex flex-col gap-1 text-sm text-gray-700 dark:text-dark-300;
}
</style>
