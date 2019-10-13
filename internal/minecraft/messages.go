package minecraft

import (
	"regexp"
	"strings"
)

var (
	// https://minecraft.gamepedia.com/Health#Death_messages
	//
	// TODO: Can we union these all together for more efficient matching?
	deathMessageMatchers = []*regexp.Regexp{
		regexp.MustCompile(`\A\S+ was shot by arrow\z`),
		regexp.MustCompile(`\A\S+ was shot by .*\z`),
		regexp.MustCompile(`\A\S+ was shot by .* using .*\z`),
		regexp.MustCompile(`\A\S+ was pricked to death\z`),
		regexp.MustCompile(`\A\S+ walked into a cactus whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ was stabbed to death\z`),
		regexp.MustCompile(`\A\S+ was roasted in dragon breath\z`),
		regexp.MustCompile(`\A\S+ was roasted in dragon breath by .*\z`),
		regexp.MustCompile(`\A\S+ drowned`),
		regexp.MustCompile(`\A\S+ drowned whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ suffocated in a wall`),
		regexp.MustCompile(`\A\S+ suffocated in a wall whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ was squished too much`),
		regexp.MustCompile(`\A\S+ was squashed by .*\z`),
		regexp.MustCompile(`\A\S+ experienced kinetic energy\z`),
		regexp.MustCompile(`\A\S+ experienced kinetic energy whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ removed an elytra while flying whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ blew up\z`),
		regexp.MustCompile(`\A\S+ was blown up by .*\z`),
		regexp.MustCompile(`\A\S+ was blown up by .* using .*\z`),
		regexp.MustCompile(`\A\S+ was killed by .*\z`),
		regexp.MustCompile(`\A\S+ hit the ground too hard\z`),
		regexp.MustCompile(`\A\S+ hit the ground too hard whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ fell from a high place\z`),
		regexp.MustCompile(`\A\S+ fell off a ladder\z`),
		regexp.MustCompile(`\A\S+ fell off some vines\z`),
		regexp.MustCompile(`\A\S+ fell out of the water\z`),
		regexp.MustCompile(`\A\S+ fell into a patch of fire\z`),
		regexp.MustCompile(`\A\S+ fell into a patch of cacti\z`),
		regexp.MustCompile(`\A\S+ was doomed to fall\z`),
		regexp.MustCompile(`\A\S+ was doomed to fall by .*\z`),
		regexp.MustCompile(`\A\S+ was doomed to fall by .* using .*\z`),
		regexp.MustCompile(`\A\S+ fell too far and was finished by .*\z`),
		regexp.MustCompile(`\A\S+ fell too far and was finished by .* using .*\z`),
		regexp.MustCompile(`\A\S+ was shot off some vines by .*\z`),
		regexp.MustCompile(`\A\S+ was shot off a ladder by .*\z`),
		regexp.MustCompile(`\A\S+ was blown from a high place by .*\z`),
		regexp.MustCompile(`\A\S+ was squashed by a falling anvil\z`),
		regexp.MustCompile(`\A\S+ was squashed by a falling anvil whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ was squashed by a falling block\z`),
		regexp.MustCompile(`\A\S+ was squashed by a falling block whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ was killed by magic\z`),
		regexp.MustCompile(`\A\S+ went up in flames\z`),
		regexp.MustCompile(`\A\S+ burned to death\z`),
		regexp.MustCompile(`\A\S+ was burnt to a crisp whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ walked into fire whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ went off with a bang\z`),
		regexp.MustCompile(`\A\S+ went off with a bang whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ tried to swim in lava\z`),
		regexp.MustCompile(`\A\S+ tried to swim in lava to escape .*\z`),
		regexp.MustCompile(`\A\S+ was struck by lightning\z`),
		regexp.MustCompile(`\A\S+ was struck by lightning whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ discovered the floor was lava\z`),
		regexp.MustCompile(`\A\S+ walked into danger zone due to .*\z`),
		regexp.MustCompile(`\A\S+ was slain by .*\z`),
		regexp.MustCompile(`\A\S+ was slain by .* using .*\z`),
		regexp.MustCompile(`\A\S+ got finished off by .*\z`),
		regexp.MustCompile(`\A\S+ got finished off by .* using .*\z`),
		regexp.MustCompile(`\A\S+ was fireballed by .*\z`),
		regexp.MustCompile(`\A\S+ was fireballed by .* using .*\z`),
		regexp.MustCompile(`\A\S+ was stung to death\z`),
		regexp.MustCompile(`\A\S+ was killed by magic\z`),
		regexp.MustCompile(`\A\S+ was killed by even more magic\z`),
		regexp.MustCompile(`\A\S+ was killed by .* using magic\z`),
		regexp.MustCompile(`\A\S+ was killed by .* using .*\z`),
		regexp.MustCompile(`\A\S+ starved to death\z`),
		regexp.MustCompile(`\A\S+ was poked to death by a sweet berry bush\z`),
		regexp.MustCompile(`\A\S+ was poked to death by a sweet berry bush whilst trying to escape .*\z`),
		regexp.MustCompile(`\A\S+ was killed trying to hurt .*\z`),
		regexp.MustCompile(`\A\S+ was killed by .* trying to hurt .*\z`),
		regexp.MustCompile(`\A\S+ was impaled by .*\z`),
		regexp.MustCompile(`\A\S+ was impaled by .* with .*\z`),
		regexp.MustCompile(`\A\S+ fell out of the world\z`),
		regexp.MustCompile(`\A\S+ fell from a high place and fell out of the world\z`),
		regexp.MustCompile(`\A\S+ didn't want to live in the same world as .*\z`),
		regexp.MustCompile(`\A\S+ withered away\z`),
		regexp.MustCompile(`\A\S+ withered away whilst fighting .*\z`),
		regexp.MustCompile(`\A\S+ was pummeled by .*\z`),
		regexp.MustCompile(`\A\S+ was pummeled by .* using .*\z`),
		regexp.MustCompile(`\A\S+ died\z`),
		regexp.MustCompile(`\A\S+ died because of .*\z`),
	}
)

func IsDeathMessage(msg string) bool {
	for _, re := range deathMessageMatchers {
		if re.MatchString(msg) {
			return true
		}
	}

	return false
}

var (
	joinLeaveMessageMatchers = []*regexp.Regexp{
		regexp.MustCompile(`\A\S+ joined the game\z`),
		regexp.MustCompile(`\A\S+ lost connection: .*\z`),
		regexp.MustCompile(`\A\S+ left the game\z`),
	}
)

func IsJoinLeaveMessage(msg string) bool {
	for _, re := range joinLeaveMessageMatchers {
		if re.MatchString(msg) {
			return true
		}
	}

	return false
}

func IsChatMessage(msg string) bool {
	return strings.HasPrefix(msg, "<")
}

var (
	advancementMessageMatchers = []*regexp.Regexp{
		regexp.MustCompile(`\A\S+ has made the advancement .*\z`),
		regexp.MustCompile(`\A\S+ has completed the challenge .*\z`),
		regexp.MustCompile(`\A\S+ has reached the goal .*\z`),
	}
)

func IsAdvancementMessage(msg string) bool {
	for _, re := range advancementMessageMatchers {
		if re.MatchString(msg) {
			return true
		}
	}

	return false
}
