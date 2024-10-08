package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/cxz66666/zju-ical/pkg/ical"
	"github.com/cxz66666/zju-ical/pkg/zjuservice"
	"github.com/cxz66666/zju-ical/pkg/zjuservice/zjuconst"
	"strings"

	"github.com/rs/zerolog/log"
)

func firstMatchTerm(configs []zjuconst.TermConfig, target zjuconst.ClassYearAndTerm) int {
	for index, item := range configs {
		if item.Term == target.Term && item.Year == target.Year {
			return index
		}
	}
	return -1
}

func GetClassCalendar(ctx context.Context, username, password string, isGRS bool) (ical.VCalendar, error) {
	var zs zjuservice.IZJUService

	ctx = context.WithValue(ctx, zjuconst.ScheduleCtxKey, zjuconst.GetConfig())

	if !isGRS {
		zs = newUgrsService(ctx)
		log.Ctx(ctx).Info().Msgf("%s is using UGRS", username)
	} else {
		if strings.HasPrefix(username, "3") {
			zs = newGrsService(ctx, true)
		} else {
			zs = newGrsService(ctx, false)
		}
	}

	if err := zs.Login(username, password); err != nil {
		return ical.VCalendar{}, err
	}

	termConfigs := zs.GetTermConfigs()
	tweaks := zs.GetTweaks()

	vCal := ical.VCalendar{}

	for _, item := range zs.GetClassTerms() {
		index := firstMatchTerm(termConfigs, item)
		if index == -1 {
			return ical.VCalendar{}, errors.New("配置文件错误，未找到指定学期的具体配置")
		}
		classOutline, err := zs.GetClassTimeTable(item.Year, item.Term, username)
		if err != nil {
			return ical.VCalendar{}, err
		}
		log.Ctx(ctx).Info().Msgf("generating class vevents %s-%s", item.Year, zjuconst.ClassTermToDescriptionString(item.Term))
		// classes to events
		list := zjuconst.ClassToVEvents(classOutline, termConfigs[index], tweaks)
		vCal.VEvents = append(vCal.VEvents, list...)
		log.Ctx(ctx).Info().Msgf("generated class vevents %s-%s", item.Year, zjuconst.ClassTermToDescriptionString(item.Term))
	}
	log.Ctx(ctx).Info().Msgf("get class vCal success ")

	// TODO cache
	return vCal, nil
}

func GetExamCalendar(ctx context.Context, username, password string, isGRS bool) (ical.VCalendar, error) {
	var zs zjuservice.IZJUService

	ctx = context.WithValue(ctx, zjuconst.ScheduleCtxKey, zjuconst.GetConfig())

	if !isGRS {
		zs = newUgrsService(ctx)
		log.Ctx(ctx).Info().Msgf("%s is using UGRS", username)
	} else {
		if strings.HasPrefix(username, "3") {
			zs = newGrsService(ctx, true)
		} else {
			zs = newGrsService(ctx, false)
		}
	}

	if err := zs.Login(username, password); err != nil {
		return ical.VCalendar{}, err
	}

	vCal := ical.VCalendar{}

	for _, item := range zs.GetExamTerms() {
		examOutline, err := zs.GetExamInfo(item.Year, item.Term, username)
		if err != nil {
			return ical.VCalendar{}, err
		}
		log.Ctx(ctx).Info().Msgf("generating exam vevents %s-%s", item.Year, zjuconst.ExamTermToDescriptionString(item.Term))
		// exam to events
		for _, exam := range examOutline {
			vCal.VEvents = append(vCal.VEvents, exam.ToVEventList()...)
		}
		log.Ctx(ctx).Info().Msgf("generated exam vevents %s-%s", item.Year, zjuconst.ExamTermToDescriptionString(item.Term))
	}
	log.Ctx(ctx).Info().Msgf("get exam vCal success")

	// TODO cache
	return vCal, nil
}

func GetBothCalendar(ctx context.Context, username, password string, isGRS bool) (ical.VCalendar, error) {
	var zs zjuservice.IZJUService

	ctx = context.WithValue(ctx, zjuconst.ScheduleCtxKey, zjuconst.GetConfig())

	if !isGRS {
		zs = newUgrsService(ctx)
		log.Ctx(ctx).Info().Msgf("%s is using UGRS", username)
	} else {
		if strings.HasPrefix(username, "3") {
			zs = newGrsService(ctx, true)
		} else {
			zs = newGrsService(ctx, false)
		}
		log.Ctx(ctx).Info().Msgf("%s is using GRS", username)
	}
	if err := zs.Login(username, password); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("login failed")
		return ical.VCalendar{}, err
	}

	termConfigs := zs.GetTermConfigs()
	tweaks := zs.GetTweaks()

	vCal := ical.VCalendar{}

	for _, item := range zs.GetClassTerms() {
		index := firstMatchTerm(termConfigs, item)
		if index == -1 {
			return ical.VCalendar{}, errors.New("配置文件错误，未找到指定学期的具体配置")
		}
		classOutline, err := zs.GetClassTimeTable(item.Year, item.Term, username)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("get class vevents failed %s-%s", item.Year, zjuconst.ClassTermToDescriptionString(item.Term))
			return ical.VCalendar{}, err
		}
		log.Ctx(ctx).Info().Msgf("generating class vevents %s-%s", item.Year, zjuconst.ClassTermToDescriptionString(item.Term))
		// classes to events
		list := zjuconst.ClassToVEvents(classOutline, termConfigs[index], tweaks)
		vCal.VEvents = append(vCal.VEvents, list...)
		log.Ctx(ctx).Info().Msgf("generated class vevents %s-%s", item.Year, zjuconst.ClassTermToDescriptionString(item.Term))
	}
	log.Ctx(ctx).Info().Msgf("get class vCal success ")

	for _, item := range zs.GetExamTerms() {
		examOutline, err := zs.GetExamInfo(item.Year, item.Term, username)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("get exam vevents %s-%s failed", item.Year, zjuconst.ExamTermToDescriptionString(item.Term))
			return ical.VCalendar{}, err
		}
		log.Ctx(ctx).Info().Msgf("generating exam vevents %s-%s", item.Year, zjuconst.ExamTermToDescriptionString(item.Term))
		// exam to events
		for _, exam := range examOutline {
			vCal.VEvents = append(vCal.VEvents, exam.ToVEventList()...)
		}
		log.Ctx(ctx).Info().Msgf("generated exam vevents %s-%s", item.Year, zjuconst.ExamTermToDescriptionString(item.Term))
	}
	log.Ctx(ctx).Info().Msgf("get exam vCal success")
	if len(vCal.VEvents) == 0 {
		log.Ctx(ctx).Error().Msg("no events created, return error")
		var tmpStr string
		if isGRS {
			tmpStr = "研究生"
		} else {
			tmpStr = "本科生"
		}
		return vCal, errors.New(fmt.Sprintf("未能生成任何日程，请确认您正在使用的账号 %s 选择了 %s 的课程，如果以上信息确认无误，请稍后重试", username, tmpStr))
	}
	// TODO cache
	return vCal, nil

}

func GetScoreCalendar(ctx context.Context, username, password string, isGRS bool) (ical.VCalendar, error) {
	var zs zjuservice.IZJUService

	if !isGRS {
		zs = newUgrsService(ctx)
		log.Ctx(ctx).Info().Msgf("%s is using UGRS", username)
	} else {
		if strings.HasPrefix(username, "3") {
			zs = newGrsService(ctx, true)
		} else {
			zs = newGrsService(ctx, false)
		}
	}

	if err := zs.Login(username, password); err != nil {
		return ical.VCalendar{}, err
	}

	vCal := ical.VCalendar{}
	scores, err := zs.GetScoreInfo(username)
	if err != nil {
		return ical.VCalendar{}, err
	}
	// cleanup 1. remove “弃修” and "缓考" and "缺考" 2. use best score for same className
	scores = zjuconst.ScoresCleanUp(scores)

	log.Ctx(ctx).Info().Msgf("generating score vevents")
	// score to events
	vevent, err := zjuconst.ScoresToVEventList(scores)
	if err != nil {
		return ical.VCalendar{}, err
	}
	vCal.VEvents = append(vCal.VEvents, vevent...)
	log.Ctx(ctx).Info().Msgf("get score vCal success")

	return vCal, nil
}
