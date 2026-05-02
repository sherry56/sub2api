<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-6xl space-y-6">
      <div class="card overflow-hidden">
        <div class="p-5 sm:p-6">
          <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-start">
            <div>
              <div class="flex flex-wrap gap-2">
                <span class="rounded-full bg-primary-100 px-3 py-1 text-xs font-semibold text-primary-800 dark:bg-primary-900/40 dark:text-primary-100">
                  {{ t('researchDrawing.userHero.badges.research') }}
                </span>
                <span class="rounded-full bg-primary-100 px-3 py-1 text-xs font-semibold text-primary-800 dark:bg-primary-900/40 dark:text-primary-100">
                  {{ t('researchDrawing.userHero.badges.sci') }}
                </span>
                <span class="rounded-full bg-primary-100 px-3 py-1 text-xs font-semibold text-primary-800 dark:bg-primary-900/40 dark:text-primary-100">
                  {{ t('researchDrawing.userHero.badges.pretrained') }}
                </span>
              </div>

              <p class="mt-4 text-xl font-semibold leading-8 text-primary-800 dark:text-primary-200 sm:text-2xl sm:leading-9">
                {{ t('researchDrawing.userHero.subtitle') }}
              </p>
              <p class="mt-3 max-w-3xl text-sm leading-6 text-slate-600 dark:text-dark-300">
                {{ t('researchDrawing.userHero.description') }}
              </p>
            </div>

            <div v-if="isAdmin" class="w-fit rounded-lg bg-primary-100 px-3 py-2 shadow-sm dark:bg-primary-900/40 lg:justify-self-end">
              <p class="text-xs font-semibold text-primary-800 dark:text-primary-100">{{ t('researchDrawing.userHero.price') }}</p>
            </div>
          </div>

        </div>
      </div>

      <section class="card space-y-3 p-4">
        <h4 class="text-base font-semibold text-gray-900 dark:text-white">
          {{ t('researchDrawing.examplesTitle') }}
        </h4>
        <div class="grid gap-3 md:grid-cols-3">
          <article
            v-for="item in exampleCards"
            :key="item.title"
            class="overflow-hidden rounded-lg border border-gray-100 bg-white transition hover:border-primary-200 hover:shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-900"
          >
            <div class="aspect-[16/9] bg-gray-50 dark:bg-dark-900">
              <img
                class="h-full w-full object-contain p-2"
                :src="item.image"
                :alt="item.title"
                loading="lazy"
                decoding="async"
              />
            </div>
            <div class="px-3 py-2.5">
              <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</h5>
              <p v-if="item.desc" class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ item.desc }}</p>
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

          <section class="space-y-4 rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
            <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.sections.generation') }}</h5>
            <div class="grid gap-4 lg:grid-cols-3">
              <label class="field-wrap">
                <span>{{ t('researchDrawing.input.generationMode') }}</span>
                <select v-model="generationInput.generationMode" class="input">
                  <option value="default">{{ t('researchDrawing.input.defaultMode') }}</option>
                  <option value="custom">{{ t('researchDrawing.input.customMode') }}</option>
                </select>
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
                <span
                  v-if="form.research_drawing_image_gen_model_name === GPT_IMAGE_2_MODEL"
                  class="text-xs text-amber-600 dark:text-amber-300"
                >
                  GPT Image 2 将使用 GPT_API_KEY 和 GPT_BASE_URL。
                </span>
              </label>
            </div>

            <p
              v-if="generationInput.generationMode === 'custom'"
              class="rounded-lg bg-primary-50 p-3 text-sm text-primary-700 dark:bg-primary-950/30 dark:text-primary-300"
            >
              {{ t('researchDrawing.input.customModeHint') }}
            </p>

            <div v-if="isCustomGenerationMode" class="space-y-4">
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
                  <span>{{ t('researchDrawing.labels.maxRefineResolution') }}</span>
                  <select v-model="form.research_drawing_max_refine_resolution" class="input">
                    <option value="2K">2K</option>
                    <option value="4K">4K</option>
                  </select>
                </label>
              </div>
            </div>
          </section>

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
            <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.progressTitle') }}</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.progressDesc') }}</p>
          </div>

          <dl class="grid grid-cols-2 gap-3 text-sm">
            <div v-if="isAdmin" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.unitPrice') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ unitPriceText }}</dd>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.input.generationMode') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">
                {{ isCustomGenerationMode ? t('researchDrawing.input.customMode') : t('researchDrawing.input.defaultMode') }}
              </dd>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900">
              <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('researchDrawing.run.quotaNeed') }}</dt>
              <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ quotaNeed }}</dd>
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

          <div class="rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between gap-3">
              <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.statusTitle') }}</h5>
              <span
                class="rounded-full px-2 py-1 text-xs font-semibold"
                :class="runStatusTone"
              >
                {{ runJobStatus?.status || t('researchDrawing.run.statuses.idle') }}
              </span>
            </div>
            <div class="mt-3 h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="runJobStatus?.status === 'error' ? 'bg-red-500' : 'bg-primary-500'"
                :style="{ width: `${runProgressPercent}%` }"
              ></div>
            </div>
            <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ runStatusText }}</p>
            <p v-if="runJobId" class="mt-2 break-all text-xs text-gray-400 dark:text-dark-500">
              {{ t('researchDrawing.run.jobId') }}：{{ runJobId }}
            </p>
          </div>

          <p
            class="rounded-lg border border-dashed p-3 text-sm leading-6"
            :class="runPreviewStarted
              ? 'border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-900 dark:bg-primary-950/30 dark:text-primary-300'
              : 'border-gray-200 bg-gray-50 text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400'"
          >
            {{ runPreviewStarted ? t('researchDrawing.run.previewStatus') : t('researchDrawing.run.idleStatus') }}
          </p>
          <section class="space-y-4 border-t border-gray-100 pt-5 dark:border-dark-700">
          <div class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200">
            {{ t('researchDrawing.run.saveHint') }}
          </div>

        <div
          v-if="selectedPreviewImage"
          class="w-full overflow-hidden rounded-lg border border-gray-100 bg-white dark:border-dark-700 dark:bg-dark-950"
        >
          <button
            class="block w-full cursor-zoom-in"
            type="button"
            :aria-label="t('researchDrawing.run.openLargePreview')"
            @click="openLargePreview(selectedPreviewImage)"
          >
          <span class="block aspect-[16/10] bg-gray-50 p-2 sm:p-3 dark:bg-dark-900">
            <img
              class="h-full w-full object-contain"
              :src="selectedPreviewImage.url"
              :alt="t('researchDrawing.run.resultAlt', { id: selectedPreviewImage.candidateId + 1 })"
            />
          </span>
          </button>
          <div class="flex flex-col gap-3 border-t border-gray-100 p-3 text-sm dark:border-dark-700">
            <div class="min-w-0 space-y-1">
              <p class="font-semibold text-gray-900 dark:text-white">
                {{ t('researchDrawing.run.candidateLabel', { id: selectedPreviewImage.candidateId + 1 }) }}
              </p>
              <p class="break-all text-xs text-gray-500 dark:text-dark-400">
                {{ t('researchDrawing.run.generatedAt') }}：{{ formatGeneratedAt(selectedPreviewImage.generatedAt) }}
                <span class="mx-1">/</span>
                {{ t('researchDrawing.run.jobId') }}：{{ selectedPreviewImage.jobId }}
                <span class="mx-1">/</span>
                候选图 ID：{{ selectedPreviewImage.candidateId }}
              </p>
            </div>
            <button
              class="btn btn-primary w-full justify-center"
              type="button"
              :disabled="downloadingImageKey === getResultImageKey(selectedPreviewImage)"
              @click="downloadResultImage(selectedPreviewImage)"
            >
              {{ downloadingImageKey === getResultImageKey(selectedPreviewImage) ? t('common.processing') : t('researchDrawing.run.download2k') }}
            </button>
          </div>
        </div>

        <div
          v-else
          class="rounded-lg border border-dashed border-gray-200 bg-gray-50 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400"
        >
          <p v-if="runImageLoading">{{ t('researchDrawing.run.loadingImages') }}</p>
          <p v-else>{{ t('researchDrawing.run.emptyResults') }}</p>
        </div>

        <div v-if="runHistoryImages.length" class="space-y-2">
          <div class="flex items-center justify-between gap-3">
            <h5 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('researchDrawing.run.sessionHistoryTitle') }}</h5>
            <span class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('researchDrawing.run.sessionHistoryLimit', { count: RUN_HISTORY_LIMIT }) }}
            </span>
          </div>
          <div class="flex gap-3 overflow-x-auto pb-1">
            <article
              v-for="image in runHistoryImages"
              :key="getResultImageKey(image)"
              class="w-48 shrink-0 overflow-hidden rounded-lg border bg-white text-left transition dark:bg-dark-950"
              :class="selectedPreviewImage && getResultImageKey(selectedPreviewImage) === getResultImageKey(image)
                ? 'border-primary-400 shadow-sm dark:border-primary-600'
                : 'border-gray-100 hover:border-primary-300 dark:border-dark-700 dark:hover:border-primary-800'"
            >
              <button class="block w-full cursor-zoom-in text-left" type="button" @click="openLargePreview(image)">
                <img
                  class="aspect-[4/3] w-full bg-gray-50 object-contain p-1.5 dark:bg-dark-900"
                  :src="image.url"
                  :alt="t('researchDrawing.run.resultAlt', { id: image.candidateId + 1 })"
                />
                <span class="block space-y-1 px-3 py-2">
                  <span class="block text-xs font-semibold text-gray-800 dark:text-dark-100">
                    {{ t('researchDrawing.run.candidateLabel', { id: image.candidateId + 1 }) }}
                  </span>
                  <span class="block text-[11px] leading-4 text-gray-500 dark:text-dark-400">
                    {{ formatGeneratedAt(image.generatedAt) }}
                  </span>
                  <span class="block truncate text-[11px] leading-4 text-gray-400 dark:text-dark-500">
                    任务 ID：{{ image.jobId }}
                  </span>
                  <span class="block text-[11px] leading-4 text-gray-400 dark:text-dark-500">
                    候选图 ID：{{ image.candidateId }}
                  </span>
                </span>
              </button>
              <div class="border-t border-gray-100 px-3 py-2 dark:border-dark-700">
                <button
                  class="btn btn-secondary w-full justify-center text-xs"
                  type="button"
                  :disabled="downloadingImageKey === getResultImageKey(image)"
                  @click.stop="downloadResultImage(image)"
                >
                  {{ downloadingImageKey === getResultImageKey(image) ? t('common.processing') : t('researchDrawing.run.download2k') }}
                </button>
              </div>
            </article>
          </div>
        </div>
          </section>
        </aside>
      </section>

      <div
        v-if="largePreviewImage"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4"
        role="dialog"
        aria-modal="true"
        @click.self="closeLargePreview"
      >
        <div class="flex max-h-full w-full max-w-6xl flex-col overflow-hidden rounded-lg bg-white shadow-xl dark:bg-dark-950">
          <div class="flex items-center justify-between gap-3 border-b border-gray-100 px-4 py-3 dark:border-dark-700">
            <div class="min-w-0">
              <p class="font-semibold text-gray-900 dark:text-white">
                {{ t('researchDrawing.run.largePreviewTitle') }}
              </p>
              <p class="truncate text-xs text-gray-500 dark:text-dark-400">
                {{ t('researchDrawing.run.jobId') }}：{{ largePreviewImage.jobId }}
                <span class="mx-1">/</span>
                候选图 ID：{{ largePreviewImage.candidateId }}
              </p>
            </div>
            <button class="btn btn-secondary" type="button" @click="closeLargePreview">
              {{ t('common.close') }}
            </button>
          </div>
          <div class="flex min-h-0 flex-1 items-center justify-center bg-gray-50 p-3 dark:bg-dark-900">
            <img
              class="max-h-[78vh] max-w-full object-contain"
              :src="largePreviewImage.url"
              :alt="t('researchDrawing.run.resultAlt', { id: largePreviewImage.candidateId + 1 })"
            />
          </div>
          <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-100 px-4 py-3 dark:border-dark-700">
            <span class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('researchDrawing.run.generatedAt') }}：{{ formatGeneratedAt(largePreviewImage.generatedAt) }}
            </span>
            <button
              class="btn btn-primary"
              type="button"
              :disabled="downloadingImageKey === getResultImageKey(largePreviewImage)"
              @click="downloadResultImage(largePreviewImage)"
            >
              {{ downloadingImageKey === getResultImageKey(largePreviewImage) ? t('common.processing') : t('researchDrawing.run.download2k') }}
            </button>
          </div>
        </div>
      </div>

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

          <div class="grid gap-4 md:grid-cols-2">
            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.gptImageBaseURL') }}</span>
              <input
                v-model="form.research_drawing_gpt_image_base_url"
                class="input"
                type="url"
                placeholder="https://api.openai.com/v1"
              />
            </label>

            <label class="field-wrap">
              <span>{{ t('researchDrawing.labels.gptImageAPIKey') }}</span>
              <input
                v-model="form.research_drawing_gpt_image_api_key"
                class="input"
                type="password"
                :placeholder="lastSettings?.research_drawing_gpt_image_api_key_configured ? t('researchDrawing.input.keepExistingSecret') : 'GPT_API_KEY'"
                autocomplete="new-password"
              />
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
              <span
                v-if="form.research_drawing_image_gen_model_name === GPT_IMAGE_2_MODEL"
                class="text-xs text-amber-600 dark:text-amber-300"
              >
                GPT Image 2 将使用 GPT_API_KEY 和 GPT_BASE_URL。
              </span>
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
  research_drawing_gpt_image_api_key: string
  research_drawing_gpt_image_base_url: string
  research_drawing_max_refine_resolution: string
  research_drawing_unit_price: number
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
  jobId: string
  candidateId: number
  paperBananaUser?: string
  generatedAt: string
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

const baseMainModelOptions = [
  {
    label: 'Gemini 3 Flash Preview',
    value: 'openrouter/google/gemini-3-flash-preview',
  },
  {
    label: 'GPT-5.5',
    value: GPT_5_5_MODEL,
  },
]

const allowedMainModelValues = new Set(baseMainModelOptions.map((option) => option.value))

const baseImageModelOptions = [
  {
    label: 'Gemini 3.1 Flash Image Preview',
    value: 'openrouter/google/gemini-3.1-flash-image-preview',
  },
  {
    label: 'GPT Image 2',
    value: GPT_IMAGE_2_MODEL,
  },
]

const allowedImageModelValues = new Set(baseImageModelOptions.map((option) => option.value))

const RESEARCH_DRAWING_DEFAULTS: ResearchDrawingForm = {
  research_drawing_exp_mode: 'demo_planner_critic',
  research_drawing_retrieval_setting: 'auto',
  research_drawing_num_candidates: 1,
  research_drawing_aspect_ratio: '16:9',
  research_drawing_max_critic_rounds: 2,
  research_drawing_main_model_name: 'openrouter/google/gemini-3-flash-preview',
  research_drawing_image_gen_model_name: 'openrouter/google/gemini-3.1-flash-image-preview',
  research_drawing_gpt_image_api_key: '',
  research_drawing_gpt_image_base_url: 'https://api.openai.com/v1',
  research_drawing_max_refine_resolution: '2K',
  research_drawing_unit_price: 2.99,
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
const runHistoryImages = ref<RunResultImage[]>([])
const selectedResultImage = ref<RunResultImage | null>(null)
const largePreviewImage = ref<RunResultImage | null>(null)
const downloadingImageKey = ref('')
let runPollTimer: number | null = null
let exampleCarouselTimer: number | null = null
const exampleSlideIndex = ref(0)
const isAdmin = computed(() => authStore.isAdmin)
const RUN_HISTORY_LIMIT = 10

const form = reactive<ResearchDrawingForm>({
  ...RESEARCH_DRAWING_DEFAULTS,
})

const mainModelOptions = computed(() => {
  return baseMainModelOptions
})

const imageModelOptions = computed(() => {
  return baseImageModelOptions
})

const generationInput = reactive<PaperBananaGenerationInput>({
  ...PAPERBANANA_INPUT_DEFAULTS,
})

const isCustomGenerationMode = computed(() => generationInput.generationMode === 'custom')

const methodOptimizationEnabled = computed(
  () => appStore.cachedPublicSettings?.research_drawing_method_optimization_enabled !== false,
)
const methodOptimizationDefaultEnabled = computed(
  () => appStore.cachedPublicSettings?.research_drawing_method_optimization_default_enabled === true,
)

const quotaNeed = computed(() => {
  const candidates = Math.max(1, Number(form.research_drawing_num_candidates) || 1)
  const criticRounds = Math.max(1, Number(form.research_drawing_max_critic_rounds) || 1)
  return candidates * (1 + criticRounds)
})

const unitPriceText = computed(() => `${Number(form.research_drawing_unit_price || 2.99).toFixed(2)} / generation`)

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

const runProgressPercent = computed(() => {
  const status = runJobStatus.value?.status
  if (status === 'done' || status === 'error') {
    return 100
  }
  if (!runJobId.value) {
    return 0
  }
  const elapsed = Math.max(0, Number(runJobStatus.value?.elapsed_sec) || 0)
  return Math.min(92, Math.max(18, 18 + elapsed * 1.2))
})

const runStatusTone = computed(() => {
  const status = runJobStatus.value?.status
  if (status === 'done') {
    return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  }
  if (status === 'error') {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  }
  if (runJobId.value) {
    return 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
  }
  return 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-dark-400'
})

const selectedPreviewImage = computed(() => selectedResultImage.value || runHistoryImages.value[0] || null)

const exampleCards = computed(() => [
  {
    title: t('researchDrawing.examples.frameworkTitle'),
    desc: '',
    image: [
      '/research-drawing/framework-structure.png',
      '/research-drawing/framework-variables.png',
    ][exampleSlideIndex.value % 2],
  },
  {
    title: t('researchDrawing.examples.workflowTitle'),
    desc: '',
    image: [
      '/research-drawing/mechanism-mediation.png',
      '/research-drawing/mechanism-moderation.png',
    ][exampleSlideIndex.value % 2],
  },
  {
    title: t('researchDrawing.examples.datasetTitle'),
    desc: '',
    image: [
      '/research-drawing/stats-bar.png',
      '/research-drawing/stats-line.png',
    ][exampleSlideIndex.value % 2],
  },
])

function startExampleCarousel() {
  if (exampleCarouselTimer !== null) {
    return
  }
  exampleCarouselTimer = window.setInterval(() => {
    exampleSlideIndex.value += 1
  }, 3500)
}

function stopExampleCarousel() {
  if (exampleCarouselTimer !== null) {
    window.clearInterval(exampleCarouselTimer)
    exampleCarouselTimer = null
  }
}

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
  form.research_drawing_gpt_image_base_url =
    form.research_drawing_gpt_image_base_url?.trim().replace(/\/+$/, '') || RESEARCH_DRAWING_DEFAULTS.research_drawing_gpt_image_base_url
  form.research_drawing_gpt_image_api_key = form.research_drawing_gpt_image_api_key?.trim() || ''
  form.research_drawing_unit_price = Math.max(
    0.01,
    Number(form.research_drawing_unit_price) || RESEARCH_DRAWING_DEFAULTS.research_drawing_unit_price,
  )
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
  form.research_drawing_gpt_image_api_key = ''
  form.research_drawing_gpt_image_base_url = settings.research_drawing_gpt_image_base_url || RESEARCH_DRAWING_DEFAULTS.research_drawing_gpt_image_base_url
  form.research_drawing_max_refine_resolution = settings.research_drawing_max_refine_resolution || RESEARCH_DRAWING_DEFAULTS.research_drawing_max_refine_resolution
  form.research_drawing_unit_price = settings.research_drawing_unit_price || RESEARCH_DRAWING_DEFAULTS.research_drawing_unit_price
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
  largePreviewImage.value = null
  runResultImages.value = []
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
  runResultImages.value = []
  stopRunPolling()
  try {
    const payload: ResearchDrawingGenerateRequest = {
      method_content: generationInput.methodContent,
      caption: generationInput.caption,
      generation_mode: generationInput.generationMode,
      optimize_method_content: isCustomGenerationMode.value ? generationInput.optimizeMethodContent : false,
      main_model_name: form.research_drawing_main_model_name,
      image_gen_model_name: form.research_drawing_image_gen_model_name,
      ...(isCustomGenerationMode.value
        ? {
            exp_mode: form.research_drawing_exp_mode,
            retrieval_setting: form.research_drawing_retrieval_setting,
            num_candidates: form.research_drawing_num_candidates,
            aspect_ratio: form.research_drawing_aspect_ratio,
            max_critic_rounds: form.research_drawing_max_critic_rounds,
            max_refine_resolution: form.research_drawing_max_refine_resolution,
          }
        : {}),
    }
    const result = await researchDrawingAPI.generate(payload)
    runJobId.value = result.job_id
    runPaperBananaUser.value = result.paperbanana_user || ''
    appStore.showInfo(
      isAdmin.value
        ? t('researchDrawing.run.submittedWithCharge', { charge: result.charge })
        : t('researchDrawing.run.previewStatus'),
    )
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
  runResultImages.value = []
  try {
    const images: RunResultImage[] = []
    const generatedAt = new Date().toISOString()
    for (const candidateId of candidateIds) {
      const blob = await researchDrawingAPI.getJobImage(runJobId.value, candidateId, runPaperBananaUser.value)
      images.push({
        jobId: runJobId.value,
        candidateId,
        paperBananaUser: runPaperBananaUser.value,
        generatedAt,
        url: URL.createObjectURL(blob),
      })
    }
    runResultImages.value = images
    appendRunHistory(images)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.run.imagesFailed')))
  } finally {
    runImageLoading.value = false
  }
}

function revokeRunImages() {
  runHistoryImages.value.forEach((image) => URL.revokeObjectURL(image.url))
  runHistoryImages.value = []
  runResultImages.value = []
  largePreviewImage.value = null
}

function openLargePreview(image: RunResultImage) {
  selectedResultImage.value = image
  largePreviewImage.value = image
}

function closeLargePreview() {
  largePreviewImage.value = null
}

function appendRunHistory(images: RunResultImage[]) {
  const next = [...images, ...runHistoryImages.value]
  const kept = next.slice(0, RUN_HISTORY_LIMIT)
  next.slice(RUN_HISTORY_LIMIT).forEach((image) => URL.revokeObjectURL(image.url))
  runHistoryImages.value = kept
  selectedResultImage.value = images[0] || kept[0] || null
}

function getResultImageKey(image: RunResultImage) {
  return `${image.jobId}:${image.candidateId}`
}

function formatGeneratedAt(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

async function downloadResultImage(image: RunResultImage) {
  const key = getResultImageKey(image)
  if (downloadingImageKey.value) {
    return
  }
  downloadingImageKey.value = key
  try {
    // TODO(research-drawing): switch this to the real PaperBanana 2K export endpoint when it is exposed.
    const blob = await researchDrawingAPI.getJobImage(image.jobId, image.candidateId, image.paperBananaUser)
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = `research-drawing-${image.jobId}-${image.candidateId}.png`
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.setTimeout(() => URL.revokeObjectURL(objectUrl), 0)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchDrawing.run.downloadFailed')))
  } finally {
    downloadingImageKey.value = ''
  }
}

function formatDuration(seconds: number) {
  const sec = Math.max(0, Math.floor(seconds))
  const min = Math.floor(sec / 60)
  const rest = sec % 60
  return min > 0 ? `${min} 分 ${rest} 秒` : `${rest} 秒`
}

function applyMethodOptimizationDefault() {
  generationInput.optimizeMethodContent =
    methodOptimizationEnabled.value && methodOptimizationDefaultEnabled.value
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
      research_drawing_gpt_image_api_key: form.research_drawing_gpt_image_api_key,
      research_drawing_gpt_image_base_url: form.research_drawing_gpt_image_base_url,
      research_drawing_max_refine_resolution: form.research_drawing_max_refine_resolution,
      research_drawing_unit_price: form.research_drawing_unit_price,
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

onMounted(async () => {
  await appStore.fetchPublicSettings()
  applyMethodOptimizationDefault()
  loadSettings()
  startExampleCarousel()
})

onBeforeUnmount(() => {
  stopExampleCarousel()
  stopRunPolling()
  revokeRunImages()
})
</script>

<style scoped>
.field-wrap {
  @apply flex flex-col gap-1 text-sm text-gray-700 dark:text-dark-300;
}
</style>
