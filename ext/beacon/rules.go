package beacon

import (
	"regexp"

	"github.com/samalba/dockerclient"
)

func (b *Beacon) ruleMatch(cfg *dockerclient.ContainerConfig) bool {
	// iterate through rules and check type / match
	isMatch := false

	for _, rule := range b.cfg.Rules {
		switch rule.Type {
		case "label":
			// TODO: match label
			m := b.isLabelMatch(rule.Regex, cfg)
			if m {
				isMatch = m
				break
			}
		case "name":
			// TODO: match name
			m := b.isNameMatch(rule.Regex, cfg)
			if m {
				isMatch = m
				break
			}
		case "image":
			// TODO: match name
			m := b.isImageMatch(rule.Regex, cfg)
			if m {
				isMatch = m
				break
			}
		default:
			log().Errorf("unknown rule type: %s", rule.Type)
		}
	}

	return isMatch
}

func (b *Beacon) isLabelMatch(rule string, cfg *dockerclient.ContainerConfig) bool {
	log().Warnf("isLabelMatch not implemented")
	return false
}

func (b *Beacon) isNameMatch(rule string, cfg *dockerclient.ContainerConfig) bool {
	log().Warnf("isNameMatch not implemented")
	return false
}

func (b *Beacon) isImageMatch(rule string, cfg *dockerclient.ContainerConfig) bool {
	image := cfg.Image

	r := regexp.MustCompile(rule)

	m := r.MatchString(image)

	return m
}
