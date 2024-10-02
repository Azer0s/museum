package impl

import (
	"museum/config"
	"museum/domain"
	"regexp"
)

var addressRegex = regexp.MustCompile(`\{\{ *@([\w-+_.]+) *}}`)
var hostRegex = regexp.MustCompile(`\{\{ *host *}}`)

type EnvironmentTemplateResolverServiceImpl struct {
	Config config.Config
}

func (s *EnvironmentTemplateResolverServiceImpl) FillEnvironmentTemplate(exhibit *domain.Exhibit, o domain.Object, _ *map[string]string) (error, map[string]string) {
	res := make(map[string]string)
	if o.Environment != nil {
		for k, v := range o.Environment {
			if v == "" {
				continue
			}

			// replace {{ @objectName }} with the object name from the containerIpMapping
			if addressRegex.MatchString(v) {
				matches := addressRegex.FindStringSubmatch(v)
				if len(matches) == 2 {
					v = addressRegex.ReplaceAllString(v, exhibit.Name+"_"+matches[1])

					/*if name, ok := (*templateContainer)[matches[1]]; ok {
						v = addressRegex.ReplaceAllString(v, name)
					} else {
						return errors.New("could not find object " + matches[1]), nil
					}*/
				}
			}

			// replace {{ host }} with the hostname and the path
			if hostRegex.MatchString(v) {
				matches := hostRegex.FindStringSubmatch(v)
				if len(matches) == 1 {
					v = hostRegex.ReplaceAllString(v, s.Config.GetHostname()+":"+s.Config.GetPort()+"/exhibit/"+exhibit.Id)
				}
			}

			res[k] = v
		}
	}

	return nil, res
}
